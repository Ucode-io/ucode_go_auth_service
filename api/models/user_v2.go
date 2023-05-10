package models

type UserId struct {
	UserId string `json:"user_id"`
}

type Language struct {
	Id         string `json:"id"`
	Name       string `json:"name"`
	ShortName  string `json:"short_name"`
	NativeName string `json:"native_name"`
}

type Timezone struct {
	Id   string `json:"id"`
	Name string `json:"name"`
	Text string `json:"text"`
}

type ListLanguage struct {
	Language []*Language `json:"language"`
	Count    int         `json:"count"`
}

type ListTimezone struct {
	Timezone []*Timezone `json:"timezone"`
	Count    int         `json:"count"`
}
