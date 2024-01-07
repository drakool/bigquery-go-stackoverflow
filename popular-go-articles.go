// Copyright 2019 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// [START bigquery_simple_app_all]

// Command simpleapp queries the Stack Overflow public dataset in Google BigQuery.
package main

// [START bigquery_simple_app_deps]
import (
	"context"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"cloud.google.com/go/bigquery"
	"google.golang.org/api/iterator"
)

func getAllPosts(c echo.Context) error {

	var window []StackOverflowRow
	window = posts[pos:windowLength]
	pos += windowLength
	return c.JSON(http.StatusOK, window)

}

func main() {

	projectID := os.Getenv("PROJECT_ID")
	if projectID == "" {
		fmt.Println("GOOGLE_CLOUD_PROJECT environment variable must be set.")
		os.Exit(1)
	}

	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.GET("/posts", getAllPosts)

	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, struct{ Status string }{Status: "OK"})
	})

	httpPort := os.Getenv("PORT")
	if httpPort == "" {
		httpPort = "8080"
	}

	// [START bigquery_simple_app_client]
	ctx := context.Background()

	client, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("bigquery.NewClient: %v", err)
	}
	defer client.Close()
	// [END bigquery_simple_app_client]

	rows, err := query(ctx, client)
	if err != nil {
		log.Fatal(err)
	}

	/*if err := printResults(os.Stdout, rows); err != nil {
		log.Fatal(err)
	}*/

	if err := getPosts(rows); err != nil {
		log.Fatal(err)
	}

	pos = 0
	windowLength = 10

	e.Logger.Fatal(e.Start(":" + httpPort))

}

// query returns a row iterator suitable for reading query results.
func query(ctx context.Context, client *bigquery.Client) (*bigquery.RowIterator, error) {

	// [START bigquery_simple_app_query]
	query := client.Query(
		`SELECT
			title,
			CONCAT(
				'https://stackoverflow.com/questions/',
				CAST(id as STRING)) as url,
			view_count,
			q.score, 
			q.answer_count as answer_count,
			q.creation_date creation_date
		FROM ` + "`bigquery-public-data.stackoverflow.posts_questions` as q" + `
		WHERE tags like '%go%' and 
			q.comment_count > 10 and
			q.creation_date between timestamp(DATE_SUB(current_date(), INTERVAL 2 year)) and timestamp(current_date())
		ORDER BY view_count DESC
		LIMIT 100;`)
	return query.Read(ctx)
	// [END bigquery_simple_app_query]
}

// [START bigquery_simple_app_print]
type StackOverflowRow struct {
	Title        string    `bigquery:"title"`
	URL          string    `bigquery:"url"`
	ViewCount    int64     `bigquery:"view_count"`
	Score        int64     `bigquery:"score"`
	AnswerCount  int64     `bigquery:"answer_count"`
	CreationDate time.Time `bigquery:"creation_date"`
}

var (
	posts        []StackOverflowRow
	pos          int
	windowLength int
)

func getPosts(iter *bigquery.RowIterator) error {

	for {
		var row StackOverflowRow
		err := iter.Next(&row)

		if err == iterator.Done {
			return nil
		}
		if err != nil {
			return fmt.Errorf("error iterating through results: %w", err)
		}
		posts = append(posts, row)
		//fmt.Fprintf(w, "title: %s createion-date: %v url: %s views: %d\n", row.Title, row.CreationDate, row.URL, row.ViewCount)
		//fmt.Fprintf(w, "title: %s createion-date: %v url: %s views: %d\n", row.Title, row.CreationDate, row.URL, row.ViewCount)
	}
}

// printResults prints results from a query to the Stack Overflow public dataset.
func printResults(w io.Writer, iter *bigquery.RowIterator) error {
	for {
		var row StackOverflowRow
		err := iter.Next(&row)
		if err == iterator.Done {
			return nil
		}
		if err != nil {
			return fmt.Errorf("error iterating through results: %w", err)
		}
		fmt.Fprintf(w, "title: %s createion-date: %v url: %s views: %d\n", row.Title, row.CreationDate, row.URL, row.ViewCount)
		//fmt.Fprintf(w, "title: %s createion-date: %v url: %s views: %d\n", row.Title, row.CreationDate, row.URL, row.ViewCount)
	}
}

// [END bigquery_simple_app_print]
// [END bigquery_simple_app_all]
