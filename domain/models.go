package domain

type Subject struct {
	ID           int           `json:"subject_id" gorm:"primaryKey;column:subject_id"`
	Name         string        `json:"subject_name" gorm:"column:name"`
	MBTI         string        `json:"mbti" gorm:"column:mbti"`
	SubjectTypes []SubjectType `json:"subject_types" gorm:"foreignKey:SubjectID"`
	ImageURL     string        `json:"image_url" gorm:"column:image_url"`
	GroupID      *int          `json:"group_id" gorm:"foreignKey:GroupID;constraint:OnDelete:SET NULL"`
}

type Typology struct {
	TypologyID  int    `json:"typology_id" gorm:"primaryKey;column:typology_id"`
	Name        string `json:"typology_name" gorm:"column:name"`
	DisplayName string `json:"typology_display_name" gorm:"column:display_name"`
}

type Type struct {
	ID          int    `json:"type_id" gorm:"primaryKey;column:type_id;unique"`
	TypologyID  int    `json:"typology_id" gorm:"primaryKey;column:typology_id"`
	Name        string `json:"type_name" gorm:"column:name"`
	DisplayName string `json:"type_display_name" gorm:"column:display_name"`
	Description string `json:"type_description" gorm:"column:description"`
}

type TypeForSubject struct {
	ID int `json:"type_id" gorm:"primaryKey;column:type_id"`
}

type SubjectType struct {
	SubjectID  int      `gorm:"primaryKey;column:subject_id"`
	TypologyID int      `gorm:"primaryKey;column:typology_id"`
	TypeID     int      `gorm:"column:type_id"`
	Subject    Subject  `gorm:"foreignKey:SubjectID"`
	Typology   Typology `gorm:"foreignKey:TypologyID"`
	Type       Type     `gorm:"foreignKey:TypeID"`
}

type User struct {
	UserID   int    `gorm:"primaryKey;column:user_id"`
	Username string `gorm:"column:user_name;unique"`
	PassHash string `gorm:"column:pass_hash"`
}

type Group struct {
	GroupID   int    `json:"group_id" gorm:"primaryKey;column:group_id"`
	Groupname string `json:"group_name" gorm:"column:group_name;unique"`
}

type SubjectResponse struct {
	Subject   string `json:"subject"`
	SubjectID int    `json:"subject_id"`
	Types     []int  `json:"types"`
}
