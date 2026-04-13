module github.com/savvinovan/kyc-service

go 1.25

require (
	github.com/go-chi/chi/v5 v5.2.5
	github.com/ilyakaznacheev/cleanenv v1.5.0
	github.com/savvinovan/event-sourcing-learning/contracts v0.0.0
	go.uber.org/fx v1.24.0
)

require (
	github.com/BurntSushi/toml v1.2.1 // indirect
	github.com/joho/godotenv v1.5.1 // indirect
	go.uber.org/dig v1.19.0 // indirect
	go.uber.org/multierr v1.10.0 // indirect
	go.uber.org/zap v1.26.0 // indirect
	golang.org/x/sys v0.0.0-20220412211240-33da011f77ad // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	olympos.io/encoding/edn v0.0.0-20201019073823-d3554ca0b0a3 // indirect
)

replace github.com/savvinovan/event-sourcing-learning/contracts => ../contracts
