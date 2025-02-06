package meals

type User struct {
	Id       int
	Password string
	Email    string
}

type UserService interface {
	Get(id int) (User, error)
	GetFromSessionId(sessionId string) (User, error)
	GetFromCredentials(email, password string) (User, error)
	GetFromEmail(email string) (User, error)
	Add(user User) (User, error)
	UpdateEmail(user User, email string) error
	UpdatePassword(user User, password string) error
}
