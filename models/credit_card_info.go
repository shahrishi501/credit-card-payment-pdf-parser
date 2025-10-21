package models


type CreditCardInfo struct {
	CardVariant string
	CardLast4Digits string
	BillingCycle string
	PaymentDueDate string
	TotalBalance string
	Transactions []Transaction
}

type Transaction struct {
	ID     string
	Amount string
	TransactionDetail string
	Date   string
}