package main

import (
	"database/sql"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
	//"gopkg.in/yaml.v2"
	"html/template"
	//"io/ioutil"
	//"regexp"
	//"strconv"
	"strings"
	"bufio"
	"bytes"
)

const postsFile = "data/posts.yml"
const tweetsFile = "data/tweets.yml"
const tagsFile = "data/tags.yml"

type Paste struct {
	Expiration string
	Content    []byte
	UUID       string
}

func main() {
	r := mux.NewRouter()
	// Routes consist of a path and a handler function.
	r.HandleFunc("/", IndexHandler)
	r.HandleFunc("/posts/{id}", PostHandler)
	r.HandleFunc("/new", NewPostHandler)
	r.HandleFunc("/analytics", AnalyticsHandler)


	// Bind to a port and pass our router in
	log.Fatal(http.ListenAndServe(":8003", r))
}


type YouTubeData struct {
	Id    string
	Body  string
	Music []string
}

type Post struct {
	Id          int
	Slug        string
	Title       string
	Description string
	Date        string
	TopResult   string
	Keywords    []string
	YouTubeData YouTubeData
	Body        string
	Transcript  string
	ClickToTweet string
}

func IndexHandler(w http.ResponseWriter, r *http.Request) {

	db, err := sql.Open("sqlite3", "./db/dtp.db")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	rows, err := db.Query("SELECT id, title FROM posts ORDER BY id DESC")
	if err != nil {
		panic(err)
	}

	posts := []Post{}
	for rows.Next() {
		post := new(Post)

		err = rows.Scan(&post.Id, &post.Title)
		if err != nil {
			panic(err)
		}
		posts = append(posts, *post)
	}

	//---- Page data
	data := struct {
		Title string
		Posts []Post
	}{
		"Homey McHomePage",
		posts,
	}

	t, err := template.ParseFiles("templates/index.html")
	if err != nil {
		log.Fatal(err)
	}

	err = t.Execute(w, data)
	if err != nil {
		log.Fatal(err)
	}
}


func NewPostHandler(w http.ResponseWriter, r *http.Request) {

	//@todo - refactor database connection
	db, err := sql.Open("sqlite3", "./db/dtp.db")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	//vars := mux.Vars(r)
	//id := vars["id"]

	r.ParseForm()


	if r.Method=="POST"{

		//-------   POST   ------

		stmt, err := db.Prepare("INSERT INTO posts (id, slug, title, description, body) VALUES (?, ?, ?, ?, ?)" )
		if err != nil {
			panic(err)
		}

		_, err = stmt.Exec(r.FormValue("id"), r.FormValue("slug"), "Dummy title", "Dummy description", "Dummy body")
		if err != nil {
			log.Fatal(err)
		}

		//----   YOUTUBE
		stmt, err = db.Prepare("INSERT INTO youtube (id, post_id, body) VALUES (?, ?, ?)" )
		if err != nil {
			panic(err)
		}
		dummyId := fmt.Sprintf("%s-DUMMY",  r.FormValue("id"))
		_, err = stmt.Exec(dummyId, r.FormValue("id"), "Dummy body")
		if err != nil {
			log.Fatal(err)
		}

		//Redirect to the end page
		route := fmt.Sprintf("/posts/%s",  r.FormValue("id"))
		http.Redirect(w, r, route, 301)
	}


	t, err := template.ParseFiles("templates/new.html")
	if err != nil {
		log.Fatal(err)
	}
	err = t.Execute(w, nil)
	if err != nil {
		log.Fatal(err)
	}
}



func PostHandler(w http.ResponseWriter, r *http.Request) {

	//@todo - refactor database connection
	db, err := sql.Open("sqlite3", "./db/dtp.db")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	vars := mux.Vars(r)
	id := vars["id"]

	r.ParseForm()


	if r.Method=="POST"{

		//-------   POST   ------

		stmtUpdate, err := db.Prepare("UPDATE posts SET slug = ?, title = ?, description = ?, topresult = ?, click_to_tweet = ?, body = ? , transcript = ? WHERE id = ?" )
		if err != nil {
			panic(err)
		}
		_, err = stmtUpdate.Exec(r.FormValue("slug"), r.FormValue("title"), r.FormValue("description"),  r.FormValue("top_result"), r.FormValue("click_to_tweet"), r.FormValue("body"), r.FormValue("transcript"), id)
		if err != nil {
			log.Fatal(err)
		}

		//YouTube
		stmtUpdateYT, err := db.Prepare("UPDATE youtube SET id = ?, body = ? WHERE post_id = ?" )
		if err != nil {
			panic(err)
		}
		_, err = stmtUpdateYT.Exec(r.FormValue("yt_id"), r.FormValue("yt_body"), id)
		if err != nil {
			log.Fatal(err)
		}

		//-------   KEYWORDS   ------
		//DELETE
		stmt, err := db.Prepare("DELETE FROM posts_keywords_xref WHERE post_id = ?" )
		if err != nil {
			panic(err)
		}
		_, err = stmt.Exec(id)
		if err != nil {
			log.Fatal(err)
		}

		//INSERT
		stmt2, err := db.Prepare("INSERT INTO  posts_keywords_xref (post_id, keyword_id, sort_order) VALUES (?, ?, ?)" )
		if err != nil {
			panic(err)
		}
		reader := bytes.NewReader([]byte(r.FormValue("keywords")))
		scanner := bufio.NewScanner(reader)

		i := 0
		for scanner.Scan() {
			i++
			keyword := scanner.Text()
			keyword = strings.TrimSpace(keyword)
			if keyword == "" {
				continue
			}
			_, err = stmt2.Exec(id, keyword, i)
			if err != nil {
				//do nothing: dupe
			}
		}
	}

	//POST
	rows, err := db.Query("SELECT p.id, p.slug, p.title, p.description, p.body,  COALESCE(p.click_to_tweet, '') AS click_to_tweet,   COALESCE(p.topresult, '') AS topresult,   COALESCE(p.transcript, '') AS transcript,    yt.id, yt.body FROM posts AS p  LEFT JOIN youtube AS yt ON p.id = yt.post_id WHERE p.id = ?", id)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	post := new(Post)
	if rows.Next() {
		err = rows.Scan(&post.Id, &post.Slug, &post.Title, &post.Description, &post.Body,  &post.ClickToTweet, &post.TopResult, &post.Transcript, &post.YouTubeData.Id, &post.YouTubeData.Body)
		if err != nil {
			panic(err)
		}
	}


	//Keywords
	rows2, err := db.Query("SELECT keyword_id FROM posts_keywords_xref WHERE post_id = ?  ORDER BY sort_order", id)
	if err != nil {
		panic(err)
	}
	defer rows2.Close()

	keywords := []string{}
	for rows2.Next() {

		keyword := ""
		err = rows2.Scan(&keyword)
		if err != nil {
			panic(err)
		}
		keywords = append(keywords, keyword)
	}

	//---- Page data
	data := struct {
		Post Post
		Keywords []string
	}{
		*post,
		keywords,
	}

	t, err := template.ParseFiles("templates/post.html")
	if err != nil {
		log.Fatal(err)
	}

	err = t.Execute(w, data)
	if err != nil {
		log.Fatal(err)
	}
}



func AnalyticsHandler(w http.ResponseWriter, r *http.Request) {

	//@todo - refactor database connection
	db, err := sql.Open("sqlite3", "./db/dtp.db")
	if err != nil {
		panic(err)
	}
	defer db.Close()


	//Too many keywords
	rows, err := db.Query("SELECT p.id, p.title FROM posts AS p LEFT JOIN posts_keywords_xref kw ON p.id = kw.post_id GROUP BY p.id HAVING COUNT(kw.keyword_id) > 8 ORDER BY COUNT(kw.keyword_id) DESC")
	if err != nil {
		panic(err)
	}
	defer rows.Close()


	posts := []Post{}
	for rows.Next() {
		post := new(Post)
		err = rows.Scan(&post.Id, &post.Title)
		if err != nil {
			panic(err)
		}
		posts = append(posts, *post)
	}

	//---- Page data
	data := struct {
		Posts []Post
	}{
		posts,
	}

	t, err := template.ParseFiles("templates/analytics.html")
	if err != nil {
		log.Fatal(err)
	}

	err = t.Execute(w, data)
	if err != nil {
		log.Fatal(err)
	}
}


//
//
//func populateTags(db *sql.DB) {
//
//	stmt, err := db.Prepare("INSERT INTO posts_keywords_xref (post_id, keyword_id) VALUES (?,?)")
//	if err != nil {
//		panic(err)
//	}
//
//	postTags := getTags(tagsFile)
//
//	for k, tags := range postTags {
//		postId, _ := splitSlug(k)
//		for _, tag := range tags.Tags {
//
//			if string(tag) == "Gary Straughan" {
//				continue
//			}
//
//			if string(tag) == "Development That Pays" {
//				continue
//			}
//
//			_, err := stmt.Exec(postId, string(tag))
//			if err != nil {
//				panic(err)
//			}
//		}
//	}
//}
//
//func populateTweets(db *sql.DB) {
//
//	stmt, err := db.Prepare("UPDATE posts SET click_to_tweet=? WHERE id=?")
//	if err != nil {
//		panic(err)
//	}
//
//	tweets := getTweets(tweetsFile)
//
//	for k, tweet := range tweets {
//
//		postId, _ := splitSlug(k)
//
//		_, err := stmt.Exec(tweet.Link, postId)
//		if err != nil {
//			panic(err)
//		}
//	}
//}
//
//func populatePosts(db *sql.DB) {
//
//	posts := getPosts(
//		postsFile)
//
//	stmt, err := db.Prepare("INSERT INTO posts (id, slug, title, description, published, body, transcript, topresult) values(?,?,?,?,?,?,?,?)")
//	if err != nil {
//		panic(err)
//	}
//
//	//stmtKeywords, err := db.Prepare("INSERT INTO posts_keywords_xref (post_id, keyword_id) values(?,?)")
//	//if err != nil {
//	//	panic(err)
//	//}
//
//	stmtYouTube, err := db.Prepare("INSERT INTO youtube (id, post_id, body) values(?,?,?)")
//	if err != nil {
//		panic(err)
//	}
//
//	stmtYouTubeMusic, err := db.Prepare("INSERT INTO youtube_music_xref (youtube_id, music_id) values(?,?)")
//	if err != nil {
//		panic(err)
//	}
//
//	for k, post := range posts {
//
//		id, slug := splitSlug(k)
//
//		_, err := stmt.Exec(id, slug, post.Title, post.Description, post.Date, post.Body, post.Transcript, post.TopResult)
//		if err != nil {
//			panic(err)
//		}
//
//		////Keywords
//		//_, err = stmtKeywords.Exec(id, post.Keyword)
//		//if err != nil {
//		//	panic(err)
//		//}
//
//		//Youtube
//		_, err = stmtYouTube.Exec(post.YouTubeData.Id, id, post.YouTubeData.Body)
//		if err != nil {
//			panic(err)
//		}
//
//		//YouTube Music
//		for _, music := range post.YouTubeData.Music {
//			_, err = stmtYouTubeMusic.Exec(post.YouTubeData.Id, music)
//			if err != nil {
//				panic(err)
//			}
//		}
//	}
//}
//
//func splitSlug(s string) (int, string) {
//
//	re := regexp.MustCompile(`(\d+)-(.*)`)
//
//	result := re.FindAllStringSubmatch(s, -1)
//
//	indexString := result[0][1]
//	slug := result[0][2]
//
//	index, err := strconv.Atoi(indexString)
//	if err != nil {
//		panic("oops")
//	}
//
//	return index, slug
//}
//
//func generateYML(db *sql.DB) {
//
//	posts := make(map[string]Post)
//
//	rows, err := db.Query("SELECT id, slug, title, description, published, body, transcript, topresult FROM posts")
//	if err != nil {
//		panic(err)
//	}
//
//	for rows.Next() {
//		var id int
//		var slug string
//
//		p := new(Post)
//
//		err = rows.Scan(&id, &slug, &p.Title, &p.Description, &p.Date, &p.Body, &p.Transcript, &p.TopResult)
//		if err != nil {
//			panic(err)
//		}
//
//		//Tags/Keywords
//		//keywords := new(Keywords)
//
//		rows2, err := db.Query("SELECT keyword_id FROM posts_keywords_xref WHERE post_id = ?", id)
//		if err != nil {
//			panic(err)
//		}
//
//		for rows2.Next() {
//
//			keyword := ""
//
//			err = rows2.Scan(&keyword)
//			if err != nil {
//				panic(err)
//			}
//
//			p.Keywords = append(p.Keywords, keyword)
//		}
//
//		//Youtube
//		yt := new(YouTubeData)
//
//		rows3, err := db.Query("SELECT id, body FROM youtube WHERE post_id = ?", id)
//		if err != nil {
//			panic(err)
//		}
//
//		for rows3.Next() {
//			err = rows3.Scan(&yt.Id, &yt.Body)
//			if err != nil {
//				panic(err)
//			}
//
//			rows4, err := db.Query("SELECT music_id  FROM youtube_music_xref WHERE youtube_id = ?", yt.Id)
//			if err != nil {
//				panic(err)
//			}
//
//			for rows4.Next() {
//
//				var music string
//				err = rows4.Scan(&music)
//				if err != nil {
//					panic(err)
//				}
//
//				yt.Music = append(yt.Music, music)
//			}
//
//			//Assign to the post
//			p.YouTubeData = *yt
//		}
//
//		slug = fmt.Sprintf("%d-%s", id, slug)
//		posts[slug] = *p
//
//	}
//
//	toYAML(posts)
//}
//
//func getPosts(postsFle string) map[string]Post {
//
//	data := readYAMLFile(postsFle)
//	posts := convertYAML(data)
//
//	return posts
//}
//
//func readYAMLFile(filename string) []byte {
//
//	data, err := ioutil.ReadFile(filename)
//
//	if err != nil {
//		log.Fatalf("Failed to read YML file : %v", err.Error())
//	}
//
//	return data
//}

//
//func convertYAML(input []byte) map[string]Post {
//	posts := make(map[string]Post)
//
//	err := yaml.Unmarshal(input, &posts)
//	if err != nil {
//		log.Fatalf("error: %v", err)
//	}
//	return posts
//}
//
//func toYAML(posts map[string]Post) {
//
//	bytes, err := yaml.Marshal(posts)
//	if err != nil {
//		log.Fatalf("error: %v", err)
//	}
//	ioutil.WriteFile("data/out.yml", bytes, 0644)
//
//}
//
//func getTweets(tweetsFile string) map[string]Tweet {
//
//	data := readYAMLFile(tweetsFile)
//	tweets := convertTweetsYAML(data)
//
//	return tweets
//}
//
//type Tweet struct {
//	Link string
//}
//
//func convertTweetsYAML(input []byte) map[string]Tweet {
//	tweets := make(map[string]Tweet)
//
//	err := yaml.Unmarshal(input, &tweets)
//	if err != nil {
//		log.Fatalf("error: %v", err)
//	}
//	return tweets
//}
//
//func getTags(tagsFile string) map[string]Tags {
//
//	data := readYAMLFile(tagsFile)
//	tags := convertTagsYAML(data)
//
//	return tags
//}
//
//type Tag string
//
//type Tags struct {
//	Tags []Tag
//}
//
//func convertTagsYAML(input []byte) map[string]Tags {
//	tags := make(map[string]Tags)
//
//	err := yaml.Unmarshal(input, &tags)
//	if err != nil {
//		log.Fatalf("error: %v", err)
//	}
//	return tags
//}
