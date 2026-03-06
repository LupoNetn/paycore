package transfer

type CreateTransactionRequest struct {
	SenderWalletID    string `json:"sender_wallet_id" binding:"required,uuid"`
	ReceiverWalletID  string `json:"receiver_wallet_id" binding:"omitempty,uuid"`
	ReceiverAccountNo string `json:"receiver_account_no" binding:"omitempty"`
	TransactionType   string `json:"transaction_type" binding:"required,oneof=transfer deposit withdrawal"`
	Amount            string `json:"amount" binding:"required"` // Using string for precision from frontend
	Description       string `json:"description"`
	Currency          string `json:"currency" binding:"required,len=3"`
	IdempotencyKey    string `json:"idempotency_key" binding:"required"`
}
