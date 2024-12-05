package main
import (
	"fmt"
	"strconv"
	"database/sql"
	"encoding/json"
	"net/http"
	"os"




)

type Song struct {
	ID          int json:"id"
	Group       string json:"group"
	SongName    string json:"song"
	ReleaseDate string json:"releaseDAte"
	Text        string json:"text"
	Link        string json:"link"
}

// Подключение к бд
func connectDB() (*sql.DB, error) {
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv(DB_NAME)


    // Формирование строки подключения
    connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", dbHost, dbPort, dbUser, dbPassword, dbName)

    // Подключение к бд
    db ,err := sql.Open("postgres", connStr)
    if err != nil {
	    return nil, err
    }

	return db, nil
}

// Получение списка песен с фильтрацией и пагинацией
func getSong(c *gin.Context) {
	db ,err := connectDB
	if err != nil {
		logrus.Errorf("Error connecting to database: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error connecting to database"})
        return
	}
	defer db.Close()

	// Получение параметров фильтрации и пагинации из запроса
	group := c.Query("group")
	song := c.Query("song")
	page, _ := strconv.Atoi(c.Query("page"))
	limit, _ := strconv.Atoi(c.Query("limit"))

	// Формирование SQL запроса
    query := "SELECT * FROM songs WHERE 1=1"
    if group != "" {
		query += " AND group_name = ?"
	}
	if song != "" {
		query += " AND song_name = ?"
	}
	query += " ORDER BY id LIMIT ? OFFSET ?"
    // Выполнение SQL запроса
	rows, err := db.Query(query, group, song, limit, (page-1)*limit)
	if err != nil {
		logrus.Errorf("Error executing query: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error executing query"})
		return
	}
	defer rows.Close()
	// Формирование списка песен
	var songs []Song
	for rows.Next() {
		var s Song
		if err := rows.Scan(&s.ID, &s.Group, &s.SongName, &s.ReleaseDate, &s.Text, &s.Link); err != nil {
			logrus.Errorf("Error scanning row: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error scanning row"})
			return
		}
		songs = append(songs, s)
	}

	// Возврат списка песен
	c.JSON(http.StatusOK, songs)
}

// Получение текста песни с пагинацией по куплетам
func getSongText(c *gin.Context) {
	db, err := connectDB()
	if err != nil {
	 logrus.Errorf("Error connecting to database: %v", err)
	 c.JSON(http.StatusInternalServerError, gin.H{"error": "Error connecting to database"})
	 return
	}
	defer db.Close()
	// Получение ID песни из запроса
	songID, _ := strconv.Atoi(c.Param("id"))

	// Получение текста песни с пагинацией
	page, _ := strconv.Atoi(c.Query("page"))
	limit, _ := strconv.Atoi(c.Query("limit"))

	// Формирование SQL запроса
	query := "SELECT text FROM songs WHERE id = ? LIMIT ? OFFSET ?"

	// Выполнение SQL запроса
	rows, err := db.Query(query, songID, limit, (page-1)*limit)
	if err != nil {
		logrus.Errorf("Error executing query: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error executing query"})
		return
	}
	defer rows.Close()

	// Формирование текста песни
	var text string
	for rows.Next() {
		var t string
		if err := rows.Scan(&t); err != nil {
			logrus.Errorf("Error scanning row: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error scanning row"})
			return
		}
		text += t
	}
   
	// Возврат текста песни
	c.JSON(http.StatusOK, gin.H{"text": text})
}

// Удаление песни
func deleteSong(c *gin.Context) {
	db, err := connectDB()
	if err != nil {
		logrus.Errorf("Error connecting to database: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error connecting to database"})
		return
	}
	defer db.Close()

	// Получение ID песни из запроса
	songID, _ := strconv.Atoi(c.Param("id"))

	// Формирование SQL запроса
	query := "DELETE FROM songs WHERE id = ?"
   
	// Выполнение SQL запроса
	_, err = db.Exec(query, songID)
	if err != nil {
		logrus.Errorf("Error executing query: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error executing query"})
		return
	}
   
	// Возврат успешного ответа
	c.JSON(http.StatusOK, gin.H{"message": "Song deleted successfully"})
}

// Изменение данных песни
func updateSong(c *gin.Context) {
	db, err := connectDB()
	if err != nil {
		logrus.Errorf("Error connecting to database: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error connecting to database"})
		return
	}
	defer db.Close()
   
	// Получение ID песни из запроса
	songID, _ := strconv.Atoi(c.Param("id"))
   
	// Получение данных песни из тела запроса
	var song Song
	if err := c.ShouldBindJSON(&song); err != nil {
		logrus.Errorf("Error binding JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
   
	// Формирование SQL запроса
	query := "UPDATE songs SET group_name = ?, song_name = ?, release_date = ?, text = ?, link = ? WHERE id = ?"
   
	// Выполнение SQL запроса
	_, err = db.Exec(query, song.Group, song.SongName, song.ReleaseDate, song.Text, song.Link, songID)
	if err != nil {
		logrus.Errorf("Error executing query: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error executing query"})
		return
	}
   
	// Возврат обновленных данных песни
	c.JSON(http.StatusOK, song)
}

// Добавление новой песни
func createSong(c *gin.Context) {
	db, err := connectDB()
	if err != nil {
		logrus.Errorf("Error connecting to database: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error connecting to database"})
		return
	}
	defer db.Close()
   
	// Получение данных песни из тела запроса
	var song Song
	if err := c.ShouldBindJSON(&song); err != nil {
		logrus.Errorf("Error binding JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
   
	// Запрос в API для получения дополнительной информации о песне
	songInfo, err := getSongInfo(song.Group, song.SongName)
	if err != nil {
		logrus.Errorf("Error getting song info: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting song info"})
		return
	}
   
	// Формирование SQL запроса
	query := "INSERT INTO songs (group_name, song_name, release_date, text, link) VALUES (?, ?, ?, ?, ?)"
   
	// Выполнение SQL запроса
	_, err = db.Exec(query, song.Group, song.SongName, songInfo.ReleaseDate, songInfo.Text, songInfo.Link)
	if err != nil {
		logrus.Errorf("Error executing query: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error executing query"})
		return
	}
   
	// Возврат созданной песни
	c.JSON(http.StatusCreated, song)
}

// Получение информации о песне из внешнего API
func getSongInfo(group, song string) (*SongDetail, error) {
	// Формирование URL запроса
	url := fmt.Sprintf("%s?group=%s&song=%s", os.Getenv("SONG_INFO_API"), group, song)
   
	// Выполнение HTTP-запроса
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
   
	// Десериализация ответа
	var songDetail SongDetail
	if err := json.NewDecoder(resp.Body).Decode(&songDetail); err != nil {
		return nil, err
	}
	return &songDetail, nil
}
   

func main() {
	// Настройка маршрутизатора
	r := gin.Default()
   
	// Определение маршрутов
	r.GET("/songs", getSongs)
	r.GET("/songs/:id/text", getSongText)
	r.DELETE("/songs/:id", deleteSong)
	r.PUT("/songs/:id", updateSong)
	r.POST("/songs", createSong)
   
	// Запуск сервера
	r.Run()
}
   
   
   
