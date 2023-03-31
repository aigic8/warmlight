package db

import (
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var ErrNotFound = gorm.ErrRecordNotFound

type DB struct {
	c *gorm.DB
}

type (
	User struct {
		ID        uint `gorm:"primaryKey"`
		Quotes    []Quote
		Sources   []Source
		Tags      []Tag
		FirstName string    `gorm:"not null"`
		CreatedAt time.Time `gorm:"autoCreateTime"`
		UpdatedAt time.Time `gorm:"autoUpdateTime"`
	}

	Quote struct {
		gorm.Model
		Text       string `gorm:"not null;uniqueIndex:idx_quote,priority:2"`
		UserID     uint   `gorm:"not null;uniqueIndex:idx_quote,priority:1"`
		MainSource *string
		Sources    []Source `gorm:"many2many:quote_source;"`
		Tags       []Tag    `gorm:"many2many:quote_tag;"`
	}

	Source struct {
		gorm.Model
		Name       string  `gorm:"not null;uniqueIndex:idx_source,priority:2"`
		UserID     uint    `gorm:"not null;uniqueIndex:idx_source,priority:1"`
		MainQuotes []Quote `gorm:"foreignKey:MainSource;references:Name"`
		Quotes     []Quote `gorm:"many2many:quote_source;"`
	}

	Tag struct {
		gorm.Model
		Name   string  `gorm:"not null;uniqueIndex:idx_tag,priority:2"`
		UserID uint    `gorm:"not null;uniqueIndex:idx_tag,priority:1"`
		Quotes []Quote `gorm:"many2many:quote_tag;"`
	}
)

func NewDB(URL string) (*DB, error) {
	db, err := gorm.Open(sqlite.Open(URL), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return &DB{c: db}, nil
}

func (db *DB) Init() error {
	return db.c.AutoMigrate(&User{}, &Quote{}, &Source{}, &Tag{})
}

func (db *DB) GetUser(ID uint) (*User, error) {
	var user User
	if err := db.c.First(&user, ID).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

func (db *DB) GetOrCreateUser(ID uint, firstName string) (*User, bool, error) {
	var user User
	res := db.c.Where("ID = ?", ID).Attrs(&User{ID: ID, FirstName: firstName}).FirstOrCreate(&user)
	return &user, res.RowsAffected == 1, res.Error
}

func (db *DB) CreateQuoteWithData(userID uint, text, mainSource string, tagNames []string, sourceNames []string) (*Quote, error) {
	tags := make([]Tag, 0, len(tagNames))
	for _, name := range tagNames {
		tags = append(tags, Tag{UserID: userID, Name: name})
	}

	sources := make([]Source, 0, len(sourceNames))
	for _, name := range sourceNames {
		sources = append(sources, Source{UserID: userID, Name: name})
	}
	quote := Quote{
		UserID:     userID,
		Text:       text,
		MainSource: &mainSource,
		Sources:    sources,
		Tags:       tags,
	}

	if err := db.c.Create(&quote).Error; err != nil {
		return nil, err
	}

	if err := db.c.Save(&quote).Error; err != nil {
		return nil, err
	}

	return &quote, nil
}

func (db *DB) debugClean() error {
	// FIXME find a better way?
	if err := db.c.Where("TRUE").Delete(&Tag{}).Error; err != nil {
		return err
	}
	if err := db.c.Where("TRUE").Delete(&Quote{}).Error; err != nil {
		return err
	}
	if err := db.c.Where("TRUE").Delete(&Source{}).Error; err != nil {
		return err
	}
	if err := db.c.Where("TRUE").Delete(&User{}).Error; err != nil {
		return err
	}

	return nil
}

func (db *DB) Close() error {
	// FIXME implement!
	return nil
}
