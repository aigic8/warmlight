package db

import (
	"database/sql"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var ErrNotFound = gorm.ErrRecordNotFound

type DB struct {
	c *gorm.DB
}

type (
	User struct {
		ID                 uint   `gorm:"primaryKey"`
		ChatID             uint   `gorm:"not null"`
		FirstName          string `gorm:"not null"`
		ActiveSource       sql.NullString
		ActiveSourceExpire sql.NullTime
		Quotes             []Quote
		Sources            []Source
		Tags               []Tag
		CreatedAt          time.Time `gorm:"autoCreateTime"`
		UpdatedAt          time.Time `gorm:"autoUpdateTime"`
	}

	Quote struct {
		gorm.Model
		Text   string `gorm:"not null;uniqueIndex:idx_quote,priority:2"`
		UserID uint   `gorm:"not null;uniqueIndex:idx_quote,priority:1"`
		// FIXME convert to sql.NullString
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

	Output struct {
		gorm.Model
		ChatID   uint   `gorm:"not null"`
		UserID   uint   `gorm:"not null"`
		Title    string `gorm:"not null"`
		IsActive bool   `gorm:"default:false"`
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
	return db.c.AutoMigrate(&User{}, &Quote{}, &Source{}, &Tag{}, &Output{})
}

func (db *DB) GetUser(ID uint) (*User, error) {
	var user User
	if err := db.c.First(&user, ID).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

func (db *DB) GetOrCreateUser(ID, ChatID uint, firstName string) (*User, bool, error) {
	var user User
	res := db.c.Where("ID = ?", ID).Attrs(&User{ID: ID, ChatID: ChatID, FirstName: firstName}).FirstOrCreate(&user)
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

func (db *DB) CreateSource(userID uint, name string) (*Source, error) {
	source := Source{UserID: userID, Name: name}
	if err := db.c.Create(&source).Error; err != nil {
		return nil, err
	}
	return &source, nil
}

func (db *DB) GetSource(userID uint, name string) (*Source, error) {
	var res Source
	if err := db.c.Where(&Source{UserID: userID, Name: name}).Take(&res).Error; err != nil {
		return nil, err
	}

	return &res, nil
}

func (db *DB) SetActiveSource(userID uint, activeSourceStr string, activeSourceExpireTime time.Time) (bool, error) {
	activeSource := sql.NullString{Valid: true, String: activeSourceStr}
	activeSourceExpire := sql.NullTime{Valid: true, Time: activeSourceExpireTime}
	res := User{ID: userID}
	update := db.c.
		Model(&res).
		Updates(User{ActiveSource: activeSource, ActiveSourceExpire: activeSourceExpire})

	return update.RowsAffected == 1, update.Error
}

// deactivate user sources with expired active source and returns users with valid `firstName`, `chatID` and `ID`
func (db *DB) DeactivateExpiredSources() ([]User, error) {
	var users []User
	returningColumns := []clause.Column{{Name: "first_name"}, {Name: "chat_id"}, {Name: "id"}}
	update := db.c.
		Model(&users).
		Clauses(clause.Returning{Columns: returningColumns}).
		Where("active_source IS NOT NULL AND active_source_expire <= ?", time.Now()).
		Updates(map[string]interface{}{"active_source": sql.NullString{}, "active_source_expire": sql.NullTime{}})
	if err := update.Error; err != nil {
		return nil, err
	}

	return users, nil
}

func (db *DB) DeactivateSource(userID uint) error {
	return db.c.
		Model(&User{}).
		Where(&User{ID: userID}).
		Updates(map[string]interface{}{"active_source": sql.NullString{}, "active_source_expire": sql.NullTime{}}).
		Error
}

func (db *DB) GetOutputs(userID uint) ([]Output, error) {
	var outputs []Output
	if err := db.c.Where(&Output{UserID: userID}).Find(&outputs).Error; err != nil {
		return nil, err
	}

	return outputs, nil
}

func (db *DB) GetOutput(userID uint, chatTitle string) (*Output, error) {
	var output Output
	if err := db.c.Where(&Output{UserID: userID, Title: chatTitle}).Take(&output).Error; err != nil {
		return nil, err
	}

	return &output, nil
}

func (db *DB) SetOutputActive(userID uint, chatTitle string) error {
	return db.c.Model(&Output{}).Where(&Output{UserID: userID, Title: chatTitle}).Update("is_active", true).Error
}

func (db *DB) GetOrCreateOutput(userID uint, chatID uint, chatTitle string) (*Output, bool, error) {
	var output Output
	q := db.c.Where("user_id = ? AND title = ?", userID, chatTitle).Attrs(&Output{ChatID: chatID, Title: chatTitle, UserID: userID}).FirstOrCreate(&output)
	return &output, q.RowsAffected == 1, q.Error
}

func (db *DB) DeleteOutput(userID uint, chatTitle string) error {
	return db.c.Delete(&Output{}, "user_id = ? AND title = ?", userID, chatTitle).Error
}

func (db *DB) Close() error {
	// FIXME implement!
	return nil
}
