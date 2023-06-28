package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/atrox/haikunatorgo"
	"github.com/labstack/gommon/log"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"

	"github.com/PlanktoScope/cli/pkg/clients/planktoscope"
)

func makeClientID(instanceID string) string {
	if instanceID == "" {
		instanceID = haikunator.New().Haikunate()
	}
	return fmt.Sprintf("planktoscope/cli/%s", instanceID)
}

func makeConnectedClient(c *cli.Context) (*planktoscope.Client, planktoscope.Logger, error) {
	apiURL := c.String("api")
	clientID := makeClientID(c.String("instance-id"))
	config, err := planktoscope.GetConfig(apiURL, clientID)
	if err != nil {
		return nil, nil, errors.Wrap(err, "couldn't make MQTT client config")
	}
	logger := log.New(clientID)
	logger.SetLevel(log.Lvl(c.Uint64("log-level")))
	client, err := planktoscope.NewClient(config, logger)
	if err != nil {
		return nil, logger, errors.Wrapf(err, "couldn't make client for %s", apiURL)
	}

	logger.Infof("Connecting to %s", apiURL)
	if err := client.Connect(); err != nil {
		return client, logger, errors.Wrapf(err, "couldn't connect to %s", apiURL)
	}
	logger.Infof("Connected to %s", apiURL)
	return client, logger, nil
}

func listenAll(ctx context.Context, client *planktoscope.Client) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-client.PumpStateBroadcasted():
			fmt.Printf("%+v\n", client.GetState().Pump)
		case <-client.CameraStateBroadcasted():
			fmt.Printf("%+v\n", client.GetState().CameraSettings)
		case <-client.ImagerStateBroadcasted():
			fmt.Printf("%+v\n", client.GetState().Imager)
		}
	}
}

// listen

func devListenAction(c *cli.Context) error {
	client, logger, err := makeConnectedClient(c)
	if err != nil {
		return err
	}

	ctxRun, cancelRun := signal.NotifyContext(
		context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT,
	)
	listenAll(ctxRun, client)
	cancelRun()

	logger.Info("Closing connection...")
	err = client.Shutdown(context.Background())
	if err != nil {
		client.Close()
	}
	return nil
}

// hal listen

func listenHAL(ctx context.Context, client *planktoscope.Client) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-client.PumpStateBroadcasted():
			fmt.Printf("%+v\n", client.GetState().Pump)
		case <-client.CameraStateBroadcasted():
			fmt.Printf("%+v\n", client.GetState().CameraSettings)
		}
	}
}

func devHALListenAction(c *cli.Context) error {
	client, logger, err := makeConnectedClient(c)
	if err != nil {
		return err
	}

	ctxRun, cancelRun := signal.NotifyContext(
		context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT,
	)
	listenHAL(ctxRun, client)
	cancelRun()

	logger.Info("Closing connection...")
	err = client.Shutdown(context.Background())
	if err != nil {
		client.Close()
	}
	return nil
}

// ctl listen

func listenCtl(ctx context.Context, client *planktoscope.Client) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-client.ImagerStateBroadcasted():
			fmt.Printf("%+v\n", client.GetState().Imager)
		}
	}
}

func devCtlListenAction(c *cli.Context) error {
	client, logger, err := makeConnectedClient(c)
	if err != nil {
		return err
	}

	ctxRun, cancelRun := signal.NotifyContext(
		context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT,
	)
	listenCtl(ctxRun, client)
	cancelRun()

	logger.Info("Closing connection...")
	err = client.Shutdown(context.Background())
	if err != nil {
		client.Close()
	}
	return nil
}

// proc listen

func listenProc(ctx context.Context, client *planktoscope.Client) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-client.SegmenterStateBroadcasted():
			fmt.Printf("%+v\n", client.GetState().Segmenter)
		}
	}
}

func devProcListenAction(c *cli.Context) error {
	client, logger, err := makeConnectedClient(c)
	if err != nil {
		return err
	}

	ctxRun, cancelRun := signal.NotifyContext(
		context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT,
	)
	listenProc(ctxRun, client)
	cancelRun()

	logger.Infof("Closing connection to %s...", client.Config.URL)
	err = client.Shutdown(context.Background())
	if err != nil {
		client.Close()
	}
	return nil
}

// proc start

func devProcStartAction(c *cli.Context) error {
	client, logger, err := makeConnectedClient(c)
	if err != nil {
		return err
	}

	ctxRun, cancelRun := signal.NotifyContext(
		context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT,
	)
	if err = startProc(ctxRun, c, client, logger); err != nil {
		return errors.Wrap(err, "couldn't start data processing routine")
	}
	cancelRun()

	logger.Infof("Closing connection to %s...", client.Config.URL)
	err = client.Shutdown(context.Background())
	if err != nil {
		client.Close()
	}
	return nil
}

func startProc(
	ctx context.Context, c *cli.Context, client *planktoscope.Client, logger planktoscope.Logger,
) error {
	logger.Info("starting segmentation...")
	token, err := client.StartSegmenting(
		[]string{c.String("path")}, c.Uint64("processing-id"), c.Bool("recurse"),
		c.Bool("force-reprocessing"), c.Bool("keep-objects"), c.Bool("export-ecotaxa"),
	)
	if err != nil {
		return errors.Wrap(err, "couldn't send command to start segmenting")
	}
	if token.Wait(); token.Error() != nil {
		return token.Error()
	}

	return listenStartProc(ctx, client, logger, c.Bool("await-started"), c.Bool("await-finished"))
}

func listenStartProc(
	ctx context.Context, client *planktoscope.Client, logger planktoscope.Logger,
	awaitStarted, awaitFinished bool,
) error {
	if !awaitStarted && !awaitFinished {
		return nil
	}

	segmenting := false
	started := false
	finished := false
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-client.SegmenterStateBroadcasted():
			prevSegmenting := segmenting
			state := client.GetState().Segmenter
			logger.Debugf("State updated: %+v\n", state)
			if !state.StateKnown {
				break
			}
			segmenting = state.Segmenting
			started = !prevSegmenting && segmenting
			finished = prevSegmenting && !segmenting
			if started {
				logger.Info("Segmentation has started!")
				if awaitStarted && !awaitFinished {
					logger.Info("Quitting because segmentation started!")
					return nil
				}
			}
			if finished {
				logger.Info("Segmentation has finished!")
				if awaitFinished {
					logger.Info("Quitting because segmentation finished!")
					logger.Infof("Total segmented objects: %d\n", state.LastObject+1)
					return nil
				}
			}
		}
	}
}
