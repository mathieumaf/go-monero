package daemon

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/mathieumaf/go-monero/cmd/monero/display"
	"github.com/mathieumaf/go-monero/pkg/zmq"
)

type zmqCommand struct {
	JSON bool

	topic    string
	endpoint string
}

var zmqTopics = []string{
	string(zmq.TopicMinimalTxPoolAdd),
	string(zmq.TopicFullTxPoolAdd),
	string(zmq.TopicMinimalChainMain),
	string(zmq.TopicFullChainMain),
	string(zmq.TopicFullMinerData),
}

func (c *zmqCommand) Cmd() *cobra.Command {
	var topicChoicesTxt = fmt.Sprintf("(%s)", strings.Join(zmqTopics, ","))

	cmd := &cobra.Command{
		Use:   "zmq",
		Short: "listen for zmq notifications",
		RunE:  c.RunE,
	}

	cmd.Flags().BoolVar(&c.JSON, "json",
		false, "whether or not to output the result as json")

	cmd.Flags().StringVar(&c.topic, "topic",
		"json-minimal-txpool_add", "zmq topic to subscribe to "+
			topicChoicesTxt)
	_ = cmd.MarkFlagRequired("topic")
	_ = cmd.RegisterFlagCompletionFunc("topic", c.topicCompletion)

	cmd.Flags().StringVar(&c.endpoint, "endpoint",
		"", "zero-mq endpoint to listen for publications")
	_ = cmd.MarkFlagRequired("endpoint")

	return cmd
}

func (c *zmqCommand) topicCompletion(
	_ *cobra.Command, _ []string, _ string,
) ([]string, cobra.ShellCompDirective) {
	return zmqTopics, cobra.ShellCompDirectiveDefault
}

func (c *zmqCommand) RunE(_ *cobra.Command, _ []string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client := zmq.NewClient(c.endpoint, zmq.Topic(c.topic))
	defer client.Close()

	stream, err := client.Listen(ctx)
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}

	if err := c.consumeStream(ctx, stream); err != nil {
		return fmt.Errorf("consume stream: %w", err)
	}

	return nil
}

type zmqEvent struct {
	tx   interface{}
	err  error
	stop bool
}

func nextZMQEvent(ctx context.Context, stream *zmq.Stream) zmqEvent {
	select {
	case err := <-stream.ErrC:
		return zmqEvent{err: err, stop: true}
	case <-ctx.Done():
		return zmqEvent{err: ctx.Err(), stop: true}
	case tx := <-stream.FullChainMainC:
		return zmqEvent{tx: tx}
	case tx := <-stream.FullMinerDataC:
		return zmqEvent{tx: tx}
	case tx := <-stream.FullTxPoolAddC:
		return zmqEvent{tx: tx}
	case tx := <-stream.MinimalChainMainC:
		return zmqEvent{tx: tx}
	case tx := <-stream.MinimalTxPoolAddC:
		return zmqEvent{tx: tx}
	}
}

func (c *zmqCommand) consumeStream(
	ctx context.Context, stream *zmq.Stream,
) error {
	for {
		ev := nextZMQEvent(ctx, stream)
		if ev.stop {
			return ev.err
		}
		if ev.tx == nil {
			continue
		}
		if err := display.JSON(ev.tx); err != nil {
			return fmt.Errorf("display json: %w", err)
		}
	}
}

func init() {
	RootCommand.AddCommand((&zmqCommand{}).Cmd())
}
