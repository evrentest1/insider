package deliverystatus

var (
	Pending    = newType("Pending")
	InProgress = newType("InProgress")
	Failed     = newType("Failed")
	Success    = newType("Success")
)

var deliveryStatuses = make(map[string]DeliveryStatus)

type DeliveryStatus struct {
	value string
}

func newType(deliveryStatus string) DeliveryStatus {
	ht := DeliveryStatus{deliveryStatus}
	deliveryStatuses[deliveryStatus] = ht
	return ht
}

func (ds DeliveryStatus) String() string {
	return ds.value
}
