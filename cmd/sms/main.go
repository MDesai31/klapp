// sms sends a one-off text message, e.g. to a worker. It's a manual CLI
// tool for now — not wired into the worker or admin sites.
package main

import (
	"flag"
	"fmt"
	"os"

	"klapp/internal/sms"
)

func main() {
	to := flag.String("to", "", "destination phone number, E.164 format e.g. +15551234567 (required)")
	body := flag.String("body", "", "message text (required)")
	flag.Parse()

	if *to == "" || *body == "" {
		fmt.Fprintln(os.Stderr, "usage: sms -to <+15551234567> -body <message>")
		os.Exit(1)
	}

	if err := sms.Send(*to, *body); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Printf("sent to %s\n", *to)
}
