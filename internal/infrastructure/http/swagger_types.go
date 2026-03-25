package http

type SwaggerErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type SwaggerErrorResponse struct {
	Error SwaggerErrorDetail `json:"error"`
}

type SwaggerTokenResponse struct {
	Token string `json:"token"`
}

type SwaggerDummyLoginRequest struct {
	Role string `json:"role"`
}

type SwaggerRegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

type SwaggerLoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type SwaggerCreateRoomRequest struct {
	Capacity    *int    `json:"capacity,omitempty"`
	Description *string `json:"description,omitempty"`
	Name        string  `json:"name"`
}

type SwaggerCreateScheduleRequest struct {
	DaysOfWeek []int  `json:"daysOfWeek"`
	EndTime    string `json:"endTime"`
	RoomID     string `json:"roomId"`
	StartTime  string `json:"startTime"`
}

type SwaggerCreateBookingRequest struct {
	CreateConferenceLink *bool  `json:"createConferenceLink,omitempty"`
	SlotID               string `json:"slotId"`
}

type SwaggerUser struct {
	CreatedAt string `json:"createdAt,omitempty"`
	Email     string `json:"email"`
	ID        string `json:"id"`
	Role      string `json:"role"`
}

type SwaggerRoom struct {
	Capacity    *int    `json:"capacity,omitempty"`
	CreatedAt   string  `json:"createdAt,omitempty"`
	Description *string `json:"description,omitempty"`
	ID          string  `json:"id"`
	Name        string  `json:"name"`
}

type SwaggerSchedule struct {
	DaysOfWeek []int  `json:"daysOfWeek"`
	EndTime    string `json:"endTime"`
	ID         string `json:"id,omitempty"`
	RoomID     string `json:"roomId"`
	StartTime  string `json:"startTime"`
}

type SwaggerSlot struct {
	End    string `json:"end"`
	ID     string `json:"id"`
	RoomID string `json:"roomId"`
	Start  string `json:"start"`
}

type SwaggerBooking struct {
	ConferenceLink *string `json:"conferenceLink,omitempty"`
	CreatedAt      string  `json:"createdAt,omitempty"`
	ID             string  `json:"id"`
	SlotID         string  `json:"slotId"`
	Status         string  `json:"status"`
	UserID         string  `json:"userId"`
}

type SwaggerPagination struct {
	Page     int `json:"page"`
	PageSize int `json:"pageSize"`
	Total    int `json:"total"`
}

type InfoResponse struct {
	Status string `json:"status"`
}

type RegisterResponse struct {
	User SwaggerUser `json:"user"`
}

type RoomsListResponse struct {
	Rooms []SwaggerRoom `json:"rooms"`
}

type RoomResponse struct {
	Room SwaggerRoom `json:"room"`
}

type ScheduleResponse struct {
	Schedule SwaggerSchedule `json:"schedule"`
}

type SlotsListResponse struct {
	Slots []SwaggerSlot `json:"slots"`
}

type BookingResponse struct {
	Booking SwaggerBooking `json:"booking"`
}

type BookingsListResponse struct {
	Bookings []SwaggerBooking `json:"bookings"`
}

type BookingsPageResponse struct {
	Bookings   []SwaggerBooking  `json:"bookings"`
	Pagination SwaggerPagination `json:"pagination"`
}
