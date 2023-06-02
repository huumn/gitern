package stripehelper

import (
	"database/sql"
	"fmt"
	"gitern/db"
	"gitern/pubkey"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/stripe/stripe-go/v71"
	"github.com/stripe/stripe-go/v71/customer"
	"github.com/stripe/stripe-go/v71/usagerecord"
)

var (
	STRIPE_SECRET_KEY string
)

func ReportUsageTx(accountName string, tx db.DBTX) error {
	row := tx.QueryRow(`SELECT stripe_id, stripe_status
						FROM accounts 
						WHERE name = $1`, accountName)
	var sid, sst sql.NullString
	if err := row.Scan(&sid, &sst); err != nil {
		return err
	}

	// don't report usage if we aren't a paying customer
	stripeID := sid.String
	stripeStatus := sst.String
	if stripeStatus != "paid" {
		return nil
	}

	count, err := pubkey.CountTx(accountName, tx)
	if err != nil {
		return err
	}

	cust, err := customer.Get(stripeID, nil)
	if err != nil {
		return err
	}

	if len(cust.Subscriptions.Data) == 0 || len(cust.Subscriptions.Data[0].Items.Data) == 0 {
		return fmt.Errorf("Stripe customer object lacks expected subscription data")
	}

	_, err = usagerecord.New(&stripe.UsageRecordParams{
		SubscriptionItem: stripe.String(cust.Subscriptions.Data[0].Items.Data[0].ID),
		Timestamp:        stripe.Int64(time.Now().Unix()),
		Quantity:         stripe.Int64(int64(count)),
		Action:           stripe.String(stripe.UsageRecordActionSet),
	})

	return err
}

func ReportUsage(accountName string) error {
	return ReportUsageTx(accountName, db.Conn)
}

func init() {
	godotenv.Load()

	str, exists := os.LookupEnv("STRIPE_SECRET_KEY")
	if exists {
		STRIPE_SECRET_KEY = str
	}

	if STRIPE_SECRET_KEY == "" {
		log.Fatalln("STRIPE_SECRET_KEY is not set")
	}

	stripe.Key = STRIPE_SECRET_KEY
}
