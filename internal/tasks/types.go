package tasks

const (
	TypeSendOTPEmail = "task:send_otp_email"
)

type SendOTPEmailPayload struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	OTP    string `json:"otp"`
}
