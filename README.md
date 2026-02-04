# gp-migrate-data

### Các bước cài đặt:
- Cài Golang: https://go.dev/doc/install
- Chạy lênh cài package: go mod tidy
- Chạy lệnh migrate: go run cmd/main.go
### Câu trúc file:
- component.json => data cũ của section
- json_convert.json => logic migrate
- output.json => kết quả trả về sau khi migrate
Ae có thể import data trong file output.json vào editor để test ui nhé
