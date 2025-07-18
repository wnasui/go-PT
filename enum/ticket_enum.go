package enum

type TicketStatus int
type TicketTag string

const (
	TicketStatusNormal TicketStatus = iota //0:未售，1：已售，2：已退，3:已删除
	TicketStatusSold
	TicketStatusRefund
	TicketStatusDeleted
)

// 简化为枚举表示车次，前端可以按照枚举值查询或是修改车次并显示为详细信息
const (
	TicketTagA_A TicketTag = "A" //8-9点左右 成都到河南
	TicketTagA_B TicketTag = "B" //12-13点左右 河南到成都
	TicketTagA_C TicketTag = "C" //18-19点左右 河南到杭州
)

func (s TicketStatus) String() string {
	switch s {
	case TicketStatusNormal:
		return "未售"
	case TicketStatusSold:
		return "已售"
	case TicketStatusDeleted:
		return "已删除"
	default:
		return "UNKNOWN"
	}
}

func (s TicketTag) String() string {
	switch s {
	case TicketTagA_A:
		return "A"
	case TicketTagA_B:
		return "B"
	case TicketTagA_C:
		return "C"
	default:
		return "UNKNOWN"
	}
}
