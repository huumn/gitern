package logmill

import "log"

// all this package does right now is initialize logging to not have flags
// in the future we should probably actually record logs

// to disable log flags (eg timestamp printing) for user facing error messages
// import _ "gitern/logmill"
func init() {
	log.SetFlags(0)
}
