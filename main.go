package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
)

func main()  {
	var (
		region = flag.String("r", "ap-northeast-1", "region")
		profile = flag.String("p", "default", "profile")
	)
	flag.Parse()

	cfg, error := config.LoadDefaultConfig(
		context.TODO(),
		config.WithRegion(*region),
		config.WithSharedConfigProfile(*profile),
	)

	if error != nil {
		log.Fatalf("failed to load config: %v", error)
	}

	svc := ecs.NewFromConfig(cfg)
	clusterName := selectArn("cluster", listCluster(svc))
	serviceName := selectArn("service", listService(svc, &clusterName))
	taskArn := selectArn("task", listTask(svc, &clusterName, &serviceName))
	containerName := selectArn("container", listContainer(svc, &clusterName, &taskArn))

	session, err := json.Marshal(createSession(svc, &clusterName, &taskArn, &containerName))
	if err != nil {
		log.Fatalf("failed to marshal session: %v", err)
	}

	target, err := json.Marshal(map[string]string{
		"Target": fmt.Sprintf(
			"ecs:/%s_/%s_/%s",
			clusterName,
			taskArn,
			runtimeId(svc, &clusterName, &taskArn, &containerName),
		),
	})
	if err != nil {
		log.Fatalf("failed to marshal target: %v", err)
	}

	cmd := exec.Command(
		"session-manager-plugin",
		string(session),
		cfg.Region,
		"StartSession",
		*profile,
		string(target),
		fmt.Sprintf("httpd://ssm.%s.amazonaws.com/", cfg.Region),
	)

	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func listCluster(svc *ecs.Client) func() []string {
	return func() []string {
		list, err := svc.ListClusters(context.TODO(), &ecs.ListClustersInput{})
		if err != nil {
			log.Fatalf("failed to list clusters: %v", err)
		}

		clusters, err := svc.DescribeClusters(context.TODO(), &ecs.DescribeClustersInput{Clusters: list.ClusterArns})
		if err != nil {
			log.Fatalf("failed to describe clusters: %v", err)
		}

		var names []string
		for _, v := range clusters.Clusters {
			var name string = *v.ClusterName
			names = append(names, name)
		}
		return names
	}
}

func listService(svc *ecs.Client, clusterName *string) func() []string {
	return func() []string {
		list, err := svc.ListServices(context.TODO(), &ecs.ListServicesInput{Cluster: clusterName})
		if err != nil {
			log.Fatalf("failed to list service: %v", err)
		}

		clusters, err := svc.DescribeServices(context.TODO(), &ecs.DescribeServicesInput{
			Cluster: clusterName,
			Services: list.ServiceArns,
		})
		if err != nil {
			log.Fatalf("failed to describe service: %v", err)
		}

		var names []string
		for _, v := range clusters.Services {
			var name string = *v.ServiceName
			names = append(names, name)
		}
		return names
	}
}

func listTask(svc *ecs.Client, clusterName *string, serviceName *string) func() []string {
	return func() []string {
		list, err := svc.ListTasks(context.TODO(), &ecs.ListTasksInput{
			Cluster: clusterName,
			ServiceName: serviceName,
		})
		if err != nil {
			log.Fatalf("failed to list clusters: %v", err)
		}
		return list.TaskArns
	}
}

func listContainer(svc *ecs.Client, clusterName *string, taskArn *string) func() []string {
	return func() []string {
		list := describeTasks(svc, clusterName, taskArn)
		var names []string
		for _, v := range list.Tasks[0].Containers {
			var name string = *v.Name
			names = append(names, name)
		}
		return names
	}
}

func createSession(svc *ecs.Client, clusterName *string, taskArn *string, containerName *string) *types.Session {
	var ecsCommand = "/bin/bash"
	commandResponse, err := svc.ExecuteCommand(context.TODO(), &ecs.ExecuteCommandInput{
		Interactive: true,
		Cluster: clusterName,
		Task: taskArn,
		Container: containerName,
		Command: &ecsCommand,
	})
	if err != nil {
		log.Fatalf("failed to execute command: %v", err)
	}

	fmt.Println("\nstart interactive session:")
	return commandResponse.Session
}

func runtimeId(svc *ecs.Client, clusterName *string, taskArn *string, containerName *string) string {
	list := describeTasks(svc, clusterName, taskArn)

	var runtimeId string
	for _, v := range list.Tasks[0].Containers {
		if v.Name == containerName {
			runtimeId = *v.RuntimeId
			break
		}
	}
	return runtimeId
}

func describeTasks(svc *ecs.Client, clusterName *string, taskArn *string) *ecs.DescribeTasksOutput{
	list, err := svc.DescribeTasks (context.TODO(), &ecs.DescribeTasksInput{
		Cluster: clusterName,
		Tasks: []string{*taskArn},
	})
	if err != nil {
		log.Fatalf("failed to list clusters: %v", err)
	}
	return list
}

func selectArn(target string, fn func() []string) string {
	arns := fn()

	fmt.Printf("\nchoose %s:\n", target)
	for i, v := range arns {
		fmt.Printf("%d: %s\n", i+1, v)
	}
	fmt.Print("input number: ")

	intValue := 1
	if len(arns) == 1 {
		fmt.Println("1")
	} else {
		intValue = scanInt()
	}

	if (intValue <= 0 || (intValue - 1) >= len(arns)) {
		log.Fatalf("invalid number: %d", intValue)
	}

	return arns[intValue - 1]
}

func scanInt() int {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()

	intValue, err := strconv.Atoi(scanner.Text())
	if err != nil {
		log.Fatalf("failed to scan int: %v", err)
	}

	return intValue
}
