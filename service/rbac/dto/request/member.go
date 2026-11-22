package request

type MemberSendCodeRequest struct {
	Email string `json:"email" binding:"required,max=254"`
}
type MemberRegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=40"`
	Nickname string `json:"nickname" binding:"required,max=40"`
	Email    string `json:"email" binding:"required,max=254"`
	Password string `json:"password" binding:"required"`
	Code     string `json:"code" binding:"required,len=6"`
}
type MemberLoginRequest struct {
	Account  string `json:"account" binding:"required_without=Email,max=254"`
	Email    string `json:"email" binding:"required_without=Account,max=254"`
	Password string `json:"password" binding:"required"`
}
type MemberLogoutRequest struct{}
