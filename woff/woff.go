package main

import (
	"flag"
	"fmt"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/checkout/session"
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
		params := &stripe.CheckoutSessionParams{
			PaymentMethodTypes: stripe.StringSlice([]string{"card", "alipay"}),
			Mode:               stripe.String(string(stripe.CheckoutSessionModePayment)),
			CustomerEmail:      stripe.String("user@example.com"),
			LineItems: []*stripe.CheckoutSessionLineItemParams{{
				PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
					Currency:    stripe.String("cny"),
					ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{Name: stripe.String("payment")},
					UnitAmount:  stripe.Int64(int64(amount * 100)),
				},
				Quantity: stripe.Int64(1),
			}},
			SuccessURL: stripe.String(returnURL),
			CancelURL:  stripe.String(returnURL),
		}
		sess, err := session.New(params)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to create payment session: %s", err), http.StatusInternalServerError)
      return
		}
		http.Redirect(w, r, sess.URL, http.StatusFound)
	})
	http.ListenAndServe(*listen, woff)
}
