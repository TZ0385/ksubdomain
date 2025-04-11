package main

import (
	"bufio"
	"context"
	"math/rand"
	"os"

	core2 "github.com/boy-hack/ksubdomain/pkg/core"
	"github.com/boy-hack/ksubdomain/pkg/core/gologger"
	"github.com/boy-hack/ksubdomain/pkg/core/ns"
	"github.com/boy-hack/ksubdomain/pkg/core/options"
	"github.com/boy-hack/ksubdomain/pkg/runner"
	"github.com/boy-hack/ksubdomain/pkg/runner/outputter"
	output2 "github.com/boy-hack/ksubdomain/pkg/runner/outputter/output"
	processbar2 "github.com/boy-hack/ksubdomain/pkg/runner/processbar"
	"github.com/urfave/cli/v2"
)

var enumCommand = &cli.Command{
	Name:    string(options.EnumType),
	Aliases: []string{"e"},
	Usage:   "枚举域名",
	Flags: append(commonFlags, []cli.Flag{
		&cli.StringFlag{
			Name:     "filename",
			Aliases:  []string{"f"},
			Usage:    "字典路径",
			Required: false,
			Value:    "",
		},
		&cli.BoolFlag{
			Name:  "filter-wild",
			Usage: "过滤泛解析域名",
			Value: false,
		},

		&cli.BoolFlag{
			Name:  "ns",
			Usage: "读取域名ns记录并加入到ns解析器中",
			Value: false,
		},
	}...),
	Action: func(c *cli.Context) error {
		if c.NumFlags() == 0 {
			cli.ShowCommandHelpAndExit(c, "enum", 0)
		}
		var domains []string
		var processBar processbar2.ProcessBar = &processbar2.ScreenProcess{}
		var err error
		var domainTotal int = 0

		// handle domain
		if c.StringSlice("domain") != nil {
			domains = append(domains, c.StringSlice("domain")...)
		}
		if c.Bool("stdin") {
			scanner := bufio.NewScanner(os.Stdin)
			scanner.Split(bufio.ScanLines)
			for scanner.Scan() {
				domains = append(domains, scanner.Text())
			}
		}
		if c.Bool("skip-wild") {
			tmp := domains
			domains = []string{}
			for _, sub := range tmp {
				isWild, _ := core2.IsWildCard(sub)
				if !isWild {
					domains = append(domains, sub)
				} else {
					gologger.Infof("域名:%s 存在泛解析,已跳过", sub)
				}
			}
		}

		var subdomainDict []string
		if c.String("filename") == "" {
			subdomainDict = core2.GetDefaultSubdomainData()
		} else {
			subdomainDict, err = core2.LinesInFile(c.String("filename"))
			if err != nil {
				gologger.Fatalf("打开文件:%s 错误:%s", c.String("filename"), err.Error())
			}
		}

		//levelDict := c.String("level-dict")
		//var levelDomains []string
		//if levelDict != "" {
		//	dl, err := core2.LinesInFile(levelDict)
		//	if err != nil {
		//		gologger.Fatalf("读取domain文件失败:%s,请检查--level-dict参数\n", err.Error())
		//	}
		//	levelDomains = dl
		//} else if c.Int("level") > 2 {
		//	levelDomains = core2.GetDefaultSubNextData()
		//}

		render := make(chan string)
		go func() {
			defer close(render)
			for _, sub := range subdomainDict {
				for _, domain := range domains {
					dd := sub + "." + domain
					render <- dd
				}
			}
		}()
		domainTotal = len(subdomainDict) * len(domains)

		// 取域名的dns,加入到resolver中
		specialDns := make(map[string][]string)
		defaultResolver := options.GetResolvers(c.StringSlice("resolvers"))
		if c.Bool("ns") {
			for _, domain := range domains {
				nsServers, ips, err := ns.LookupNS(domain, defaultResolver[rand.Intn(len(defaultResolver))])
				if err != nil {
					continue
				}
				specialDns[domain] = ips
				gologger.Infof("%s ns:%v", domain, nsServers)
			}

		}
		if c.Bool("not-print") {
			processBar = nil
		}

		// 输出到屏幕
		screenWriter, err := output2.NewScreenOutput()
		if err != nil {
			gologger.Fatalf(err.Error())
		}
		var writer []outputter.Output
		writer = append(writer, screenWriter)
		if c.String("output") != "" {
			outputFile := c.String("output")
			outputType := c.String("output-type")
			wildFilterMode := c.String("wild-filter-mode")
			switch outputType {
			case "txt":
				p, err := output2.NewPlainOutput(outputFile, wildFilterMode)
				if err != nil {
					gologger.Fatalf(err.Error())
				}
				writer = append(writer, p)
			case "json":
				p := output2.NewJsonOutput(outputFile, wildFilterMode)
				writer = append(writer, p)
			case "csv":
				p := output2.NewCsvOutput(outputFile, wildFilterMode)
				writer = append(writer, p)
			default:
				gologger.Fatalf("输出类型错误:%s 暂不支持", outputType)
			}
		}

		opt := &options.Options{
			Rate:               options.Band2Rate(c.String("band")),
			Domain:             render,
			DomainTotal:        domainTotal,
			Resolvers:          defaultResolver,
			Silent:             c.Bool("silent"),
			TimeOut:            c.Int("timeout"),
			Retry:              c.Int("retry"),
			Method:             options.VerifyType,
			Writer:             writer,
			ProcessBar:         processBar,
			SpecialResolvers:   specialDns,
			WildcardFilterMode: c.String("wild-filter-mode"),
		}
		opt.Check()
		opt.EtherInfo = options.GetDeviceConfig(c.String("eth"))
		ctx := context.Background()
		r, err := runner.New(opt)
		if err != nil {
			gologger.Fatalf("%s\n", err.Error())
			return nil
		}
		r.RunEnumeration(ctx)
		r.Close()
		return nil
	},
}
