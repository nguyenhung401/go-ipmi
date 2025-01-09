package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/bougou/go-ipmi"
	"github.com/spf13/cobra"
)

const timeFormat = time.RFC3339

// x Experimental commands.
func NewCmdX() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "x",
		Short: "x",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return initClient()
		},
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				cmd.Help()
			}
		},
		PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
			return closeClient()
		},
	}
	cmd.AddCommand(NewCmdXGetSDR())
	cmd.AddCommand(NewCmdXGetSensor())
	cmd.AddCommand(NewCmdXGetPayloadActivationStatus())
	cmd.AddCommand(NewCmdXGetDeviceGUID())
	cmd.AddCommand(NewCmdXGetSystemGUID())
	cmd.AddCommand(NewCmdXGetPEFConfig())
	cmd.AddCommand(NewCmdXGetLanConfigFor())
	cmd.AddCommand(NewCmdXGetLanConfigFull())
	cmd.AddCommand(NewCmdXGetDCMIConfig())

	return cmd
}

func NewCmdXGetSDR() *cobra.Command {
	var show bool
	var waitIntervalSec int

	cmd := &cobra.Command{
		Use:   "get-sdr",
		Short: "get-sdr",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()

			for {
				fmt.Printf("\n\nGet SDR at %s\n", time.Now().Format(timeFormat))
				res, err := client.GetSDRs(ctx)
				if err != nil {
					fmt.Printf("GetSDRs failed, err: %s", err)
					goto WAIT
				}
				fmt.Printf("GetSDRs succeeded, %d records\n", len(res))
				if show {
					fmt.Println(ipmi.FormatSDRs(res))
				}

				goto WAIT

			WAIT:
				fmt.Printf("Wait for %d seconds\n", waitIntervalSec)
				time.Sleep(time.Duration(waitIntervalSec) * time.Second)
			}
		},
	}

	cmd.PersistentFlags().BoolVarP(&show, "show", "s", false, "show table of result")
	cmd.PersistentFlags().IntVarP(&waitIntervalSec, "wait", "w", 30, "wait sleep interval sec")

	return cmd
}

func NewCmdXGetSensor() *cobra.Command {
	var show bool
	var waitIntervalSec int

	cmd := &cobra.Command{
		Use:   "get-sensor",
		Short: "get-sensor",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()

			for {
				fmt.Printf("\n\nGet Sensors at %s\n", time.Now().Format(timeFormat))
				res, err := client.GetSensors(ctx)
				if err != nil {
					fmt.Printf("GetSensors failed, err: %s", err)
					goto WAIT
				}
				fmt.Printf("GetSensors succeeded, %d records\n", len(res))
				if show {
					fmt.Println(ipmi.FormatSensors(true, res...))
				}
				goto WAIT

			WAIT:
				fmt.Printf("Wait for %d seconds\n", waitIntervalSec)
				time.Sleep(time.Duration(waitIntervalSec) * time.Second)
			}
		},
	}

	cmd.PersistentFlags().BoolVarP(&show, "show", "s", false, "show table of result")
	cmd.PersistentFlags().IntVarP(&waitIntervalSec, "wait", "w", 30, "wait sleep interval sec")

	return cmd
}

func NewCmdXGetPayloadActivationStatus() *cobra.Command {
	cmd := &cobra.Command{
		Use: "get-payload-activation-status",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) < 1 {
				fmt.Println("usage: get-payload-activation-status {payload-type}")
				return
			}

			payloadType, err := parseStringToInt64(args[0])
			if err != nil {
				fmt.Println(err)
			}

			ctx := context.Background()
			res, err := client.GetPayloadActivationStatus(ctx, ipmi.PayloadType(payloadType))
			if err != nil {
				fmt.Println(err)
			}

			fmt.Println(res.Format())
		},
	}

	return cmd
}

func NewCmdXGetSystemGUID() *cobra.Command {
	cmd := &cobra.Command{
		Use: "get-system-guid",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			res, err := client.GetSystemGUID(ctx)
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Println(res.Format())

			fmt.Println("\nDetail of GUID\n==============")
			fmt.Println()
			fmt.Println(ipmi.FormatGUIDDetails(res.GUID))
		},
	}

	return cmd
}

func NewCmdXGetDeviceGUID() *cobra.Command {
	cmd := &cobra.Command{
		Use: "get-device-guid",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			res, err := client.GetDeviceGUID(ctx)
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Println(res.Format())

			fmt.Println("\nDetail of GUID\n==============")
			fmt.Println()
			fmt.Println(ipmi.FormatGUIDDetails(res.GUID))
		},
	}

	return cmd
}

func NewCmdXGetPEFConfig() *cobra.Command {
	cmd := &cobra.Command{
		Use: "get-pef-config",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			pefConfig, err := client.GetPEFConfig(ctx)
			if err != nil {
				fmt.Println(err)
				return
			}

			fmt.Println(pefConfig.Format())
		},
	}

	return cmd
}

func NewCmdXGetLanConfigFor() *cobra.Command {
	usage := `
	get-lan-config-for [<channel number>]
	`

	cmd := &cobra.Command{
		Use:   "get-lan-config-for",
		Short: "get-lan-config-for",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) < 1 {
				CheckErr(fmt.Errorf("usage: %s", usage))
			}

			id, err := parseStringToInt64(args[0])
			if err != nil {
				CheckErr(fmt.Errorf("invalid channel number passed, err: %s", err))
			}
			channelNumber := uint8(id)

			ctx := context.Background()

			lanConfig := ipmi.LanConfig{
				IP:               &ipmi.LanConfigParam_IP{},
				SubnetMask:       &ipmi.LanConfigParam_SubnetMask{},
				DefaultGatewayIP: &ipmi.LanConfigParam_DefaultGatewayIP{},
			}

			if err := client.GetLanConfigFor(ctx, channelNumber, &lanConfig); err != nil {
				CheckErr(fmt.Errorf("GetLanConfig failed, err: %s", err))
			}

			client.Debug("Lan Config", lanConfig)

			fmt.Println(lanConfig.Format())
		},
	}
	return cmd
}

func NewCmdXGetLanConfigFull() *cobra.Command {
	usage := `
	get-lan-config-for [<channel number>]
	`

	cmd := &cobra.Command{
		Use:   "get-lan-config-full",
		Short: "get-lan-config-full",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) < 1 {
				CheckErr(fmt.Errorf("usage: %s", usage))
			}

			id, err := parseStringToInt64(args[0])
			if err != nil {
				CheckErr(fmt.Errorf("invalid channel number passed, err: %s", err))
			}
			channelNumber := uint8(id)

			ctx := context.Background()

			lanConfig, err := client.GetLanConfigFull(ctx, channelNumber)
			if err != nil {
				CheckErr(fmt.Errorf("GetLanConfig failed, err: %s", err))
			}

			client.Debug("Lan Config", lanConfig)

			fmt.Println(lanConfig.Format())
		},
	}
	return cmd

}

func NewCmdXGetDCMIConfig() *cobra.Command {
	cmd := &cobra.Command{
		Use: "get-dcmi-config",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			dcmiConfig, err := client.GetDCMIConfig(ctx)
			if err != nil {
				fmt.Println(err)
				return
			}

			fmt.Println(dcmiConfig.Format())
		},
	}

	return cmd
}
