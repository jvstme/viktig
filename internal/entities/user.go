package entities

type User struct {
	Id           int           `json:"id" gorm:"primaryKey"` // tg user id
	Interactions []Interaction `json:"-"`
}
