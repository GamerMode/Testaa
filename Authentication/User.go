package Authentication

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

func Signup(username, password, email string) User {
	return User{Username: username, Password: password, Email: email}
}

func Signin(username, password string) User {
	return User{Username: username, Password: password, Email: ""}
}
