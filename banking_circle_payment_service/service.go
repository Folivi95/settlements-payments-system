// Package banking_circle_payment_service uses the Banking Circle API to make payments.
// This package will be moved to a separate application/service in the future so it should not have any dependencies on
// the internal/ packages of the Settlements Payments System, except for domain/models which will be replicated in the new repo.
// This service will have an incoming queue from the Settlements Payments System for unprocessed payments,
// and a response queue back with payment statuses.
package banking_circle_payment_service
