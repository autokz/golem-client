package golem

import (
	"errors"
	"github.com/streadway/amqp"
	"log"
	"strconv"
	"sync"
	"time"
)

const (
	contentTypeJSON     = "text/json"
	KindDirect          = amqp.ExchangeDirect
	KindTopic           = amqp.ExchangeTopic
	KindFanout          = amqp.ExchangeFanout
	KindHeader          = amqp.ExchangeHeaders
	pingInterval        = time.Second * 3
	initTimeoutDefault  = time.Second * 60
	defaultMessageKey   = "log"
	defaultExchangeName = "log-service.logs"
	defaultExchangeKind = "x-delayed-message"
)

var (
	ErrEmptyMessageBody = errors.New("publishing message body is empty")
	ErrNilChannel       = errors.New("amqp channel is not set")
	publisher           *Publisher
	onceInit            sync.Once
	errInit             error
)

type Publisher struct {
	project  string
	service  string
	params   *Params
	exchange *Exchange
	conn     *amqp.Connection
	ch       *amqp.Channel
}

type Params struct {
	User       string
	Password   string
	Host       string
	Port       uint32
	MessageKey string
	Mandatory  bool
	Immediate  bool
}

type Exchange struct {
	Name       string
	Kind       string
	Durable    bool
	AutoDelete bool
	Internal   bool
	NoWait     bool
	Args       map[string]interface{}
}

func InitPublisher(project, service string, params *Params, exchange *Exchange) error {
	onceInit.Do(func() {
		if project == "" {
			errInit = errors.New("project name is required")
			return
		}
		if service == "" {
			errInit = errors.New("service name is required")
			return
		}

		if params == nil {
			errInit = errors.New("config is required")
			return
		}
		if exchange == nil {
			errInit = errors.New("exchange is required")
			return
		}

		if params.MessageKey == "" {
			params.MessageKey = defaultMessageKey
		}
		if exchange.Name == "" {
			exchange.Name = defaultExchangeName
		}
		if exchange.Kind == "" {
			exchange.Kind = defaultExchangeKind
		}

		publisher = &Publisher{
			project:  project,
			service:  service,
			params:   params,
			exchange: exchange,
		}

		publisher.connectInTime(initTimeoutDefault)
	})

	return errInit
}

func (p *Publisher) connectInTime(timeout time.Duration) {
	var err error

	if timeout <= pingInterval {
		timeout = initTimeoutDefault
	}

	tickerInit := time.NewTicker(timeout)
	tickerPing := time.NewTicker(pingInterval)

	for {
		select {
		case <-tickerPing.C:
			p.conn, p.ch, err = p.connect()

			switch err {
			case nil:
				go p.connectOnClose(timeout)
				return
			default:
				log.Printf("connection error=%v\n", err)
			}
		case <-tickerInit.C:
			log.Println("failed to connect to AMQP after few seconds")
			return
		}
	}
}

func (p *Publisher) connectOnClose(timeout time.Duration) {
	notify := make(chan *amqp.Error, 1)
	defer func() {
		if _, ok := <-notify; ok {
			close(notify)
		}
	}()

	p.conn.NotifyClose(notify)
	for range notify {
		p.connectInTime(timeout)
		return
	}
}

func (p Publisher) Publish(body []byte) error {
	if len(body) == 0 {
		return ErrEmptyMessageBody
	}
	if p.ch == nil {
		return ErrNilChannel
	}

	return p.ch.Publish(
		p.exchange.Name,
		p.params.MessageKey,
		p.params.Mandatory,
		p.params.Immediate,
		amqp.Publishing{
			ContentType: contentTypeJSON,
			Body:        body,
		},
	)
}

func (p Publisher) connect() (*amqp.Connection, *amqp.Channel, error) {
	var (
		ch   *amqp.Channel
		conn *amqp.Connection
		err  error
	)
	defer func() {
		if err != nil {
			muteClose(conn, ch)
		}
	}()

	conn, err = amqp.Dial("amqp://" + p.params.User + ":" + p.params.Password + "@" + p.params.Host + ":" + strconv.Itoa(int(p.params.Port)) + "/")
	if err != nil {
		return nil, nil, errors.New("failed to set connection to AMQP: " + err.Error())
	}

	ch, err = conn.Channel()
	if err != nil {
		return nil, nil, errors.New("failed to create AMQP channel: " + err.Error())
	}

	err = ch.ExchangeDeclare(
		p.exchange.Name,
		p.exchange.Kind,
		p.exchange.Durable,
		p.exchange.AutoDelete,
		p.exchange.Internal,
		p.exchange.NoWait,
		p.exchange.Args,
	)
	if err != nil {
		return nil, nil, errors.New("failed to declare exchange: " + err.Error())
	}

	return conn, ch, nil
}

func muteClose(conn *amqp.Connection, ch *amqp.Channel) {
	if conn != nil {
		_ = conn.Close()
	}
	if ch != nil {
		_ = ch.Close()
	}
}
