package login

type Request struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type Response struct {
	Token    string `json:"token"`
	Username string `json:"username"`
	Role     string `json:"role"`
}
