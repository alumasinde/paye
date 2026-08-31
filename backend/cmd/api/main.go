package main
import (
 "context"; "log"; "os"; "os/signal"; "syscall"
 "github.com/alumasinde/budget254-paye-api/internal/app"
 "github.com/alumasinde/budget254-paye-api/internal/envfile"
)
func main() {
 if err := envfile.Load(".env"); err != nil { log.Fatal(err) }
 a, err := app.NewFromEnv(); if err != nil { log.Fatal(err) }
 ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM); defer stop()
 if err := a.Run(ctx); err != nil { log.Fatal(err) }
}
