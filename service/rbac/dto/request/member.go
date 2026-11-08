package request

type MemberSendCodeRequest struct {
	Email string `json:"email" binding:"required,max=254"`
}
type MemberRegisterRequest struct {
	Email    string `json:"email" binding:"required,max=254"`
	Password string `json:"password" binding:"required"`
	Code     string `json:"code" binding:"required,len=6"`
}
type MemberLoginRequest struct {
	Email    string `json:"email" binding:"required,max=254"`
	Password string `json:"password" binding:"required"`
}
type MemberLogoutRequest struct{}
