package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"gitern/db"
	"gitern/stripehelper"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/stripe/stripe-go/v71"
	portalSession "github.com/stripe/stripe-go/v71/billingportal/session"
	"github.com/stripe/stripe-go/v71/customer"
	"github.com/stripe/stripe-go/v71/invoice"
	"github.com/stripe/stripe-go/v71/paymentmethod"
	"github.com/stripe/stripe-go/v71/sub"
	"github.com/stripe/stripe-go/v71/webhook"
)

func stripeHTTP() {
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")

	http.HandleFunc("/billing-portal", stripeBillingPortal)
	http.HandleFunc("/create-subscription", stripeCreateSubscription)
	http.HandleFunc("/retry-invoice", stripeRetryInvoice)
	http.HandleFunc("/stripe-webhook", stripeWebhook)
}

func stripeCreateCustomer(name string) (string, error) {
	params := &stripe.CustomerParams{
		Name: stripe.String(name),
	}

	c, err := customer.New(params)
	if err != nil {
		return "", err
	}

	return c.ID, err
}

func stripeIDAndStatus(accname string) (string, string, error) {
	row := db.Conn.QueryRow(`SELECT stripe_id, stripe_status
							 FROM accounts 
							 WHERE name = $1`, accname)
	var sid, sst sql.NullString
	if err := row.Scan(&sid, &sst); err != nil {
		return "", "", err
	}

	var stripeID string
	if sid.Valid {
		stripeID = sid.String
	} else {
		var err error
		stripeID, err = stripeCreateCustomer(accname)
		if err != nil || stripeID == "" {
			return "", "", err
		}
		// store the stripeID
		_, err = db.Conn.Exec("UPDATE accounts SET stripe_id = $1 where name = $2",
			stripeID, accname)
		if err != nil {
			return "", "", err
		}
	}

	stripeStatus := sst.String
	return stripeID, stripeStatus, nil
}

func stripeBillingPortalAuth(stripeID string, w http.ResponseWriter, r *http.Request) {
	params := &stripe.BillingPortalSessionParams{
		Customer:  stripe.String(stripeID),
		ReturnURL: stripe.String("https://gitern.com/account"),
	}
	s, _ := portalSession.New(params)
	http.Redirect(w, r, s.URL, http.StatusSeeOther)
}

func stripeBillingPortal(w http.ResponseWriter, r *http.Request) {
	accname, _, err := getSession(r)
	if err != nil {
		errHandler(w, http.StatusUnauthorized)
		return
	}

	stripeID, _, err := stripeIDAndStatus(accname)
	if err != nil {
		errHandlerMsg(w, http.StatusInternalServerError, "retrieving stripe status")
		return
	}

	stripeBillingPortalAuth(stripeID, w, r)
}

func stripeCreateSubscription(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		PaymentMethodID string `json:"paymentMethodId"`
		CustomerID      string `json:"customerId"`
		PriceID         string `json:"priceId"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("json.NewDecoder.Decode: %v", err)
		return
	}

	// Attach PaymentMethod
	params := &stripe.PaymentMethodAttachParams{
		Customer: stripe.String(req.CustomerID),
	}
	pm, err := paymentmethod.Attach(
		req.PaymentMethodID,
		params,
	)
	if err != nil {
		writeJSON(w, struct {
			Error error `json:"error"`
		}{err})
		return
	}

	// Update invoice settings default
	customerParams := &stripe.CustomerParams{
		InvoiceSettings: &stripe.CustomerInvoiceSettingsParams{
			DefaultPaymentMethod: stripe.String(pm.ID),
		},
	}
	c, err := customer.Update(
		req.CustomerID,
		customerParams,
	)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("customer.Update: %v %s", err, c.ID)
		return
	}

	// Create subscription
	subscriptionParams := &stripe.SubscriptionParams{
		Customer: stripe.String(req.CustomerID),
		Items: []*stripe.SubscriptionItemsParams{
			{
				Plan: stripe.String(req.PriceID),
			},
		},
	}
	subscriptionParams.AddExpand("latest_invoice.payment_intent")
	subscriptionParams.AddExpand("pending_setup_intent")

	s, err := sub.New(subscriptionParams)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("sub.New: %v", err)
		return
	}

	writeJSON(w, s)
}

func stripeRetryInvoice(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		CustomerID      string `json:"customerId"`
		PaymentMethodID string `json:"paymentMethodId"`
		InvoiceID       string `json:"invoiceId"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("json.NewDecoder.Decode: %v", err)
		return
	}

	// Attach PaymentMethod
	params := &stripe.PaymentMethodAttachParams{
		Customer: stripe.String(req.CustomerID),
	}
	pm, err := paymentmethod.Attach(
		req.PaymentMethodID,
		params,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("paymentmethod.Attach: %v %s", err, pm.ID)
		return
	}

	// Update invoice settings default
	customerParams := &stripe.CustomerParams{
		InvoiceSettings: &stripe.CustomerInvoiceSettingsParams{
			DefaultPaymentMethod: stripe.String(pm.ID),
		},
	}
	c, err := customer.Update(
		req.CustomerID,
		customerParams,
	)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("customer.Update: %v %s", err, c.ID)
		return
	}

	// Retrieve Invoice
	invoiceParams := &stripe.InvoiceParams{}
	invoiceParams.AddExpand("payment_intent")
	in, err := invoice.Get(
		req.InvoiceID,
		invoiceParams,
	)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("invoice.Get: %v", err)
		return
	}

	writeJSON(w, in)
}

func stripeWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Printf("ioutil.ReadAll: %v", err)
		return
	}

	event, err := webhook.ConstructEvent(b, r.Header.Get("Stripe-Signature"), os.Getenv("STRIPE_WEBHOOK_SECRET"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Printf("webhook.ConstructEvent: %v", err)
		return
	}

	var stripeStatus string
	switch event.Type {
	case "invoice.paid":
		stripeStatus = "paid"
		accname, ok := event.Data.Object["customer_name"].(string)
		if !ok {
			http.Error(w, "event.Data.Object[\"customer_name\"] not string", http.StatusInternalServerError)
			log.Println("event.Data.Object[\"customer_name\"] not string", event)
			return
		}

		err = stripehelper.ReportUsage(accname)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Printf("stripeReportUsage: %v", err)
			return
		}
	case "invoice.payment_failed":
		stripeStatus = "failed"
	case "customer.subscription.deleted":
		stripeStatus = ""
	default:
		log.Printf("unsupported event: %+v\n", event)
		return
	}

	stripeID, ok := event.Data.Object["customer"].(string)
	if !ok {
		http.Error(w, "event.Data.Object[\"customer\"] not string", http.StatusInternalServerError)
		log.Println("event.Data.Object[\"customer\"] not string", event)
		return
	}

	_, err = db.Conn.Exec("UPDATE accounts SET stripe_status = $1 where stripe_id = $2",
		db.NewNullString(stripeStatus), stripeID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(v); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("json.NewEncoder.Encode: %v", err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if _, err := io.Copy(w, &buf); err != nil {
		log.Printf("io.Copy: %v", err)
		return
	}
}
