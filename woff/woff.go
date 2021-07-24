package main

import (
	"flag"
	"fmt"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/paymentintent"
	"github.com/stripe/stripe-go/v72/paymentmethod"
	"net/http"
	"os"
	"strconv"
)

var listen = flag.String("l", "127.0.0.1:8080", "listen address")

var woff chi.Router

func MustLookupEnv(key string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	} else {
		panic(fmt.Sprintf("environment variable %s is required", key))
	}
}

func main() {
	flag.Parse()
	stripe.Key = MustLookupEnv("STRIPE_SECRET_KEY")
	returnURL := MustLookupEnv("RETURN_URL")

	woff = chi.NewRouter()
	woff.Use(middleware.Logger)
	woff.Get("/", func(w http.ResponseWriter, r *http.Request) {
		amount, err := strconv.ParseFloat(r.URL.Query().Get("amount"), 64)
		if err != nil {
			http.Error(w, "query parameter amount not specified or invalid", http.StatusBadRequest)
			return
		}

		pm, err := paymentmethod.New(&stripe.PaymentMethodParams{Type: stripe.String("alipay")})
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to create payment method: %s", err), http.StatusInternalServerError)
			return
		}

		pi, err := paymentintent.New(&stripe.PaymentIntentParams{
			Amount:             stripe.Int64(int64(amount * 100)),
			Currency:           stripe.String("cny"),
			PaymentMethod:      stripe.String(pm.ID),
			PaymentMethodTypes: stripe.StringSlice([]string{"alipay"}),
		})
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to create payment intent: %s", err), http.StatusInternalServerError)
			return
		}

		pi, err = paymentintent.Confirm(pi.ID, &stripe.PaymentIntentConfirmParams{
			ReturnURL: stripe.String(returnURL),
		})
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to confirm payment intent: %s", err), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, pi.NextAction.AlipayHandleRedirect.URL, http.StatusFound)
	})

	http.ListenAndServe(*listen, woff)
}
