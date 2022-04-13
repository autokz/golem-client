# golem-client

RabbitMQ client to put message into queue.

### Init publisher to send messages:

```go
package main

func main() {
	err := golem.InitPublisher(
		"golemClientTest",
		&golem.Params{
			User:     "rabbit",
			Password: "password",
			Host:     "localhost",
			Port:     5672,
		},
		&golem.Exchange{
			Name: "golemClientTest",
			Kind:  golem.KindFanout,
			AutoDelete: true,
		},
	)
	if err != nil {
		log.Fatalln(err)
	}
}
```

### Call the necessary function in your code:

```go

    golem.Info("test info message", 100)
    
    golem.Error("test error message", 409)
    
    golem.Fatal("test fatal message")

```


### Use Recover function to send panic messages and stack traces:

```go
    defer golem.Recover()
```
