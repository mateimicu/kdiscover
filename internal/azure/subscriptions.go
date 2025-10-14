package azure

// GetSubscriptions returns Azure subscription IDs to search for AKS clusters
// For now, this returns the provided subscriptions as-is
// In the future, this could be enhanced to auto-discover subscriptions
func GetSubscriptions(subscriptions []string) []string {
	if len(subscriptions) == 0 {
		// Return empty slice if no subscriptions provided
		// User should specify at least one subscription
		return []string{}
	}
	return subscriptions
}

// GetDefaultSubscriptions returns a default set of subscriptions to search
// This is empty by default since users need to specify their subscription IDs
func GetDefaultSubscriptions() []string {
	return []string{}
}