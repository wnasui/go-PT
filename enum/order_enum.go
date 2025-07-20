package enum

type OrderStatus int

const (
	OrderStatusNormal OrderStatus = iota //0:未支付，1：已支付，2：已退,3:已删除,4:生成中
	OrderStatusPaid
	OrderStatusRefunded
	OrderStatusDeleted
	OrderStatusPending
)

func (s OrderStatus) String() string {
	switch s {
	case OrderStatusNormal:
		return "未支付"
	case OrderStatusPaid:
		return "已支付"
	case OrderStatusRefunded:
		return "已退"
	case OrderStatusDeleted:
		return "已删除"
	case OrderStatusPending:
		return "生成中"
	default:
		return "UNKNOWN"
	}
}
