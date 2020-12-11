/**
*  @file
*  @copyright defined in go-stem/LICENSE
 */

package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"strings"
	"sync"
	"time"

	"github.com/scdoproject/go-stem/common"
	"github.com/scdoproject/go-stem/consensus"
	"github.com/scdoproject/go-stem/consensus/factory"
	"github.com/scdoproject/go-stem/light"
	"github.com/scdoproject/go-stem/log"
	"github.com/scdoproject/go-stem/log/comm"
	"github.com/scdoproject/go-stem/metrics"
	miner2 "github.com/scdoproject/go-stem/miner"
	"github.com/scdoproject/go-stem/monitor"
	"github.com/scdoproject/go-stem/node"
	"github.com/scdoproject/go-stem/scdo"
	"github.com/scdoproject/go-stem/scdo/lightclients"
	"github.com/spf13/cobra"
)

var (
	scdoNodeConfigFile string
	miner              string
	metricsEnableFlag  bool
	accountsConfig     string
	threads            int
	startHeight        int

	// default is full node
	lightNode bool

	//pprofPort http server port
	pprofPort uint64

	// profileSize is used to limit when need to collect profiles, set 6GB
	profileSize = uint64(1024 * 1024 * 1024 * 6)

	maxConns       = int(0)
	maxActiveConns = int(0)
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "start the node of scdo",
	Long: `usage example:
		node.exe start -c cmd\node.json
		start a node.`,

	Run: func(cmd *cobra.Command, args []string) {
		var wg sync.WaitGroup
		nCfg, err := LoadConfigFromFile(scdoNodeConfigFile, accountsConfig)
		if err != nil {
			fmt.Printf("failed to reading the config file: %s\n", err.Error())
			return
		}
		Cast(nCfg)
		if !comm.LogConfiguration.PrintLog {
			fmt.Printf("log folder: %s\n", filepath.Join(log.LogFolder, comm.LogConfiguration.DataDir))
		}
		fmt.Printf("data folder: %s\n", nCfg.BasicConfig.DataDir)

		scdoNode, err := node.New(nCfg)
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		// Create scdo service and register the service
		slog := log.GetLogger("scdo")
		lightLog := log.GetLogger("scdo-light")
		serviceContext := scdo.ServiceContext{
			DataDir: nCfg.BasicConfig.DataDir,
		}
		ctx := context.WithValue(context.Background(), "ServiceContext", serviceContext)

		var engine consensus.Engine
		if nCfg.BasicConfig.MinerAlgorithm == common.BFTEngine {
			engine, err = factory.GetBFTEngine(nCfg.SeeleConfig.CoinbasePrivateKey, nCfg.BasicConfig.DataDir)
		} else if nCfg.BasicConfig.MinerAlgorithm == common.BFTSubchainEngine {
			engine, err = factory.GetBFTSubchainEngine(nCfg.SeeleConfig.CoinbasePrivateKey, nCfg.BasicConfig.DataDir)
		} else {
			engine, err = factory.GetConsensusEngine(nCfg.BasicConfig.MinerAlgorithm, nCfg.BasicConfig.DataSetDir)
		}

		if err != nil {
			fmt.Println(err)
			return
		}

		// start pprof http server
		if pprofPort > 0 {
			go func() {
				if err := http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", pprofPort), nil); err != nil {
					fmt.Println("Failed to start pprof http server,", err)
					return
				}
			}()
		}

		if comm.LogConfiguration.IsDebug {
			go monitorPC()
		}

		if lightNode {
			lightService, err := light.NewServiceClient(ctx, nCfg, lightLog, common.LightChainDir, scdoNode.GetShardNumber(), engine)
			if err != nil {
				fmt.Println("Create light service error.", err.Error())
				return
			}

			if err := scdoNode.Register(lightService); err != nil {
				fmt.Println(err.Error())
				return
			}

			err = scdoNode.Start()
			if err != nil {
				fmt.Printf("got error when start node: %s\n", err)
				return
			}
		} else {
			// light client manager
			manager, err := lightclients.NewLightClientManager(scdoNode.GetShardNumber(), ctx, nCfg, engine)
			if err != nil {
				fmt.Printf("create light client manager failed. %s", err)
				return
			}

			// fullnode mode
			seeleService, err := scdo.NewSeeleService(ctx, nCfg, slog, engine, manager, startHeight)
			if err != nil {
				fmt.Println(err.Error())
				return
			}

			seeleService.Miner().SetThreads(threads)

			lightServerService, err := light.NewServiceServer(seeleService, nCfg, lightLog, scdoNode.GetShardNumber())
			if err != nil {
				fmt.Println("Create light server err. ", err.Error())
				return
			}

			// monitor service
			monitorService, err := monitor.NewMonitorService(seeleService, scdoNode, nCfg, slog, "Test monitor")
			if err != nil {
				fmt.Println(err.Error())
				return
			}

			services := manager.GetServices()
			services = append(services, seeleService, monitorService, lightServerService)
			for _, service := range services {
				if err := scdoNode.Register(service); err != nil {
					fmt.Println(err.Error())
					return
				}
			}

			err = scdoNode.Start()
			if maxConns > 0 {
				seeleService.P2PServer().SetMaxConnections(maxConns)
			}
			if maxActiveConns > 0 {
				seeleService.P2PServer().SetMaxActiveConnections(maxActiveConns)
			}
			if err != nil {
				fmt.Printf("got error when start node: %s\n", err)
				return
			}

			minerInfo := strings.ToLower(miner)
			if minerInfo == "start" {
				err = seeleService.Miner().Start()
				if err != nil && err != miner2.ErrMinerIsRunning {
					fmt.Println("failed to start the miner : ", err)
					return
				}
			} else if minerInfo == "stop" {
				seeleService.Miner().Stop()
			} else {
				fmt.Println("invalid miner command, must be start or stop")
				return
			}
		}

		if metricsEnableFlag {
			metrics.StartMetricsWithConfig(
				nCfg.MetricsConfig,
				slog,
				nCfg.BasicConfig.Name,
				nCfg.BasicConfig.Version,
				nCfg.P2PConfig.NetworkID,
				nCfg.SeeleConfig.Coinbase,
			)
		}

		wg.Add(1)
		wg.Wait()
	},
}

func init() {
	rootCmd.AddCommand(startCmd)

	startCmd.Flags().StringVarP(&scdoNodeConfigFile, "config", "c", "", "scdo node config file (required)")
	startCmd.MustMarkFlagRequired("config")

	startCmd.Flags().StringVarP(&miner, "miner", "m", "start", "miner start or not, [start, stop]")
	startCmd.Flags().BoolVarP(&metricsEnableFlag, "metrics", "t", false, "start metrics")
	startCmd.Flags().StringVarP(&accountsConfig, "accounts", "", "", "init accounts info")
	startCmd.Flags().IntVarP(&threads, "threads", "", 1, "miner thread value")
	startCmd.Flags().BoolVarP(&lightNode, "light", "l", false, "whether start with light mode")
	startCmd.Flags().Uint64VarP(&pprofPort, "port", "", 0, "which port pprof http server listen to")
	startCmd.Flags().IntVarP(&startHeight, "startheight", "", -1, "the block height to start from")
	startCmd.Flags().IntVarP(&maxConns, "maxConns", "", 0, "node max connections")
	startCmd.Flags().IntVarP(&maxActiveConns, "maxActiveConns", "", 0, "node max active connections")
}

func monitorPC() {
	var info runtime.MemStats
	heapDir := filepath.Join(common.GetTempFolder(), "heapProfile")
	err := os.MkdirAll(heapDir, os.ModePerm) //os.ModePerm == 0777
	if err != nil {
		fmt.Printf("failed to create folder %s: %s\n", heapDir, err)
		return
	}

	profileDir := filepath.Join(common.GetTempFolder(), "cpuProfile")
	err = os.MkdirAll(profileDir, os.ModePerm)
	if err != nil {
		fmt.Printf("failed to create folder %s: %s\n", profileDir, err)
		return
	}

	ticker := time.NewTicker(1 * time.Minute)
	for {
		select {
		case <-ticker.C:
			runtime.ReadMemStats(&info) //since ReadMemStats is up to date, we need to a heap file (most recently completed garbage collection cycle)
			if info.Alloc > profileSize {
				heapFile := filepath.Join(heapDir, fmt.Sprint("heap-", time.Now().Format("2006-01-02-15-04-05")))
				f, err := os.Create(heapFile)
				if err != nil {
					fmt.Println("monitor create heap file err:", err)
					return
				}
				err = pprof.WriteHeapProfile(f)
				if err != nil {
					fmt.Println("monitor write heap file err:", err)
					return
				}

				profileFile := filepath.Join(profileDir, fmt.Sprint("cpu-", time.Now().Format("2006-01-02-15-04-05")))
				cpuf, err := os.Create(profileFile) // if not exist, create with permision 0666(r&w)
				if err != nil {
					fmt.Println("monitor create cpu file err:", err)
					return
				}

				if err := pprof.StartCPUProfile(cpuf); err != nil {
					fmt.Println("failed to start cpu profile err:", err)
					return
				}

				time.Sleep(20 * time.Second)
				pprof.StopCPUProfile()
			}
		}
	}
}
