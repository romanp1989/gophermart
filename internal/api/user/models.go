package user

type registrationRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}
