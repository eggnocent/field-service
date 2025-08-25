package constants

type FieldScheduleStatusName string
type FieldScheduleStatus int

const (
	Available FieldScheduleStatus = 100
	Booked    FieldScheduleStatus = 200

	AvailableString FieldScheduleStatusName = "available"
	BookingString   FieldScheduleStatusName = "booking"
)

var mapFieldScheduleStatusIntToString = map[FieldScheduleStatus]FieldScheduleStatusName{
	Available: AvailableString,
	Booked:    BookingString,
}

var mapFieldScheduleStatusStringToInt = map[FieldScheduleStatusName]FieldScheduleStatus{
	AvailableString: Available,
	BookingString:   Booked,
}

func (f FieldScheduleStatus) GetStatusString() FieldScheduleStatusName {
	return mapFieldScheduleStatusIntToString[f]
}

func (f FieldScheduleStatusName) GetStatusInt() FieldScheduleStatus {
	return mapFieldScheduleStatusStringToInt[f]
}
