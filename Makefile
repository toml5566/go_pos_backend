server:
	go run main.go

gooseup1:
	cd sql/schema && goose postgres postgres://test:password@localhost:5432/test?sslmode=disable up-by-one

gooseup:
	cd sql/schema && goose postgres postgres://test:password@localhost:5432/test?sslmode=disable up

goosedown:
	cd sql/schema && goose postgres postgres://test:password@localhost:5432/test?sslmode=disable down

test:
	go test -v -cover ./...

mock:
	mockgen -package mockdb -destination internal/database/mock/store.go github.com/toml5566/go_pos_backend/internal/database Store

.PHONY: server gooseup1 gooseup goosedown test mock