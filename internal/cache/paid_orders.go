package cache

const (
	// PaidOrderPrefix is the prefix for paid order keys in the cache
	PaidOrderPrefix = "paid_order_"
)

// MarkOrderAsPaid marks an order as paid in the cache
func MarkOrderAsPaid(driver Driver, orderNo string) error {
	return driver.Set(PaidOrderPrefix+orderNo, true, 86400*7) // Keep paid status for 7 days
}

// IsOrderPaid checks if an order is marked as paid in the cache
func IsOrderPaid(driver Driver, orderNo string) bool {
	paid, ok := driver.Get(PaidOrderPrefix + orderNo)
	if !ok {
		return false
	}

	isPaid, ok := paid.(bool)
	if !ok {
		return false
	}

	return isPaid
}
