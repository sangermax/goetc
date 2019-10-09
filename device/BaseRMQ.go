package device

import (
	"github.com/streadway/amqp"
)

type MsgRecvProc func([]byte) error

type RabbitMQ struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	hasMQ   bool
}

// 初始化 参数格式：amqp://用户名:密码@地址:端口号/host
func (p *RabbitMQ) SetupRMQ(rmqAddr string) (err error) {
	if p.channel == nil {
		p.conn, err = amqp.Dial(rmqAddr)
		if err != nil {
			return err
		}

		p.channel, err = p.conn.Channel()
		if err != nil {
			return err
		}

		p.hasMQ = true
	}

	return nil
}

// 是否已经初始化
func (p *RabbitMQ) HasMQ() bool {
	return p.hasMQ

}

// 发布消息
func (p *RabbitMQ) Publish(exchange, routekey string, body []byte) (err error) {
	err = p.channel.ExchangeDeclare(exchange, "topic", true, false, false, true, nil)
	if err != nil {
		//fmt.Printf("Publish ExchangeDeclare error:%s.\r\n", err.Error())
		return err
	}

	err = p.channel.Publish(exchange, routekey, false, false, amqp.Publishing{
		ContentType: "text/plain",
		Body:        body,
	})

	if err != nil {
		//fmt.Printf("Publish Publish error:%s.\r\n", err.Error())
		return err
	}

	return nil
}

// 监听接收到的消息
func (p *RabbitMQ) Receive(exchange, routekey string, reader MsgRecvProc) (err error) {
	err = p.channel.ExchangeDeclare(exchange, "topic", true, false, false, true, nil)
	if err != nil {
		//fmt.Printf("ExchangeDeclare error:%s.\r\n", err.Error())
		return err
	}

	q, err := p.channel.QueueDeclare(routekey, true, false, false, true, nil)
	if err != nil {
		//fmt.Printf("QueueDeclare error:%s.\r\n", err.Error())
		return err
	}

	err = p.channel.QueueBind(q.Name, routekey, exchange, true, nil)
	if err != nil {
		//fmt.Printf("QueueBind error:%s.\r\n", err.Error())
		return err
	}

	msgs, err := p.channel.Consume(q.Name, "", true, false, false, false, nil)
	if err != nil {
		//fmt.Printf("Consume error:%s.\r\n", err.Error())
		return err
	}

	go func() {
		//fmt.Printf("receive:%d.\r\n", len(msgs))
		for d := range msgs {
			reader(d.Body)

		}

	}()

	return nil

}

// 关闭连接
func (p *RabbitMQ) Close() {
	p.channel.Close()
	p.conn.Close()
	p.hasMQ = false
}
