package main

import (
	"Homework/internal/db"
	"Homework/internal/domain"
	"Homework/internal/infrastructure/kafka"
	"Homework/internal/middleware"
	"Homework/internal/model"
	receiver "Homework/internal/receiver"
	"Homework/internal/repository/in_memory_cache"
	"Homework/internal/repository/postgresql"
	redis2 "Homework/internal/repository/redis"
	"Homework/internal/sender"
	"Homework/internal/server"
	"Homework/internal/service"
	"Homework/internal/storage"
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"github.com/IBM/sarama"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	flag "github.com/spf13/pflag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const port = ":9000"
const queryParamKey = "key"

func main() {
	command := os.Args[1]
	switch command {
	case "http":
		httpMain()
	default:
		mainConsole()
	}
}

func mainConsole() {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	writeCh := make(chan []string)
	readCh := make(chan []string)
	logch := make(chan string)

	orderStor, err := storage.NewOrderStorage()
	if err != nil {
		fmt.Println("не удалось подключиться к хранилищу заказов")
		return
	}
	orderServ := service.NewOrderService(&orderStor)

	pickupPointStor, err := storage.NewPickupPointStorage()
	if err != nil {
		fmt.Println("не удалось подключиться к хранилищу пвз")
		return
	}
	pickupPointServ := service.NewPickupPointService(&pickupPointStor)

	go func() {
		sig := <-sigCh
		err = pickupPointStor.Save()
		if err != nil {
			fmt.Println("Ошибка: ", err)
		}
		fmt.Println()
		fmt.Println("Получен сигнал завершения программы:", sig)
		os.Exit(0)
	}()
	go write(logch, writeCh, pickupPointServ, orderServ)
	go read(logch, readCh, pickupPointServ, orderServ)
	go logFunc(logch)

	for {
		fs := flag.NewFlagSet("command", flag.ContinueOnError)

		orderId := fs.String("ordid", "", "id for order")
		customerId := fs.String("custid", "", "id for customer")
		termKeeping := fs.String("term-keeping", "", "term keeping of order")
		listOrdersId := fs.StringSlice("sliceid", []string{}, "slice of customer id")
		limit := fs.Int("limit", 0, "last n elements")
		onlyNotIssued := fs.Bool("only-not-issued", false, "flag for not issued orders")
		pageNumber := fs.Int("page-number", 1, "number of page")
		pageSize := fs.Int("size", 5, "size of page")
		packageType := flag.String("package-type", "", "package type of order")
		weight := flag.Float32("weight", 0, "weight of order")
		price := flag.Float32("cost", 0, "price of order")
		pickupPointName := fs.String("pickup-point-name", "", "pickup point's name")
		pickupPointAddress := fs.String("address", "", "address of pickup point")
		pickupPointContactDetails := fs.String("contact-details", "", "contact details of pickup point")

		input, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			fmt.Println(err)
		}
		args := strings.Split(strings.TrimSpace(input), " ")

		for i, arg := range args {
			args[i] = strings.TrimSpace(arg)
		}

		err = fs.Parse(args)
		if err != nil {
			fmt.Println("Ошибка при парсинге флагов:", err)
			continue
		}

		if len(fs.Args()) != 1 {
			fmt.Println("указано неверное количество команд")
		}
		command := args[0]

		switch command {
		case "accept-order":
			if *orderId == "" {
				fmt.Println("не указан id заказа")
				continue
			}

			if *customerId == "" {
				fmt.Println("не указан id клиента")
				continue
			}

			if *termKeeping == "" {
				fmt.Println("не указана дата")
				continue
			}

			_, err := time.Parse("2/1/2006", *termKeeping)
			if err != nil {
				fmt.Println("неверный формат даты")
				continue
			}

			if *packageType == "" {
				fmt.Println("не указан тип упаковки")
				return
			}

			if *weight == 0 {
				fmt.Println("не указан вес упаковки")
				return
			}

			if *weight < 0 {
				fmt.Println("вес должен быть положительным")
				return
			}

			if *price == 0 {
				fmt.Println("не указана стоимость упаковки")
				return
			}

			if *weight < 0 {
				fmt.Println("стоимость должна быть положительной")
				return
			}

			writeCh <- []string{command, *orderId, *customerId, *termKeeping, *packageType, fmt.Sprintf("%f", *weight), fmt.Sprintf("%f", *price)}

		case "return-to-courier":
			if *orderId == "" {
				fmt.Println("не указан id заказа")
				continue
			}

			writeCh <- []string{command, *orderId}

		case "give-order":
			if len(*listOrdersId) == 0 {
				fmt.Println("передан пустой список заказов")
				continue
			}

			data := []string{command}
			writeCh <- append(data, *listOrdersId...)

		case "get-list-order":
			if *customerId == "" {
				fmt.Println("не указан id клиента")
				continue
			}

			if *limit < 0 {
				fmt.Println("limit должен быть положительным")
				continue
			}

			readCh <- []string{command, *customerId, strconv.Itoa(*limit), strconv.FormatBool(*onlyNotIssued)}

		case "refund-order":
			if *orderId == "" {
				fmt.Println("не указан id заказа")
				continue
			}

			if *customerId == "" {
				fmt.Println("не указан id клиента")
				continue
			}

			writeCh <- []string{command, *orderId, *customerId}

		case "get-list-refund":
			if *pageNumber < 0 {
				fmt.Println("номер страницы должен быть положительным")
				continue
			}

			if *pageSize < 0 {
				fmt.Println("размер страницы должен быть положительным")
				continue
			}

			readCh <- []string{strconv.Itoa(*pageNumber), strconv.Itoa(*pageSize)}

		case "add-pickup-point":
			if *pickupPointName == "" {
				fmt.Println("не указано название пвз")
			} else if *pickupPointAddress == "" {
				fmt.Println("не указан адрес пвз")
			} else if *pickupPointContactDetails == "" {
				fmt.Println("не указаны контактные данные пвз")
			} else {
				writeCh <- []string{command, *pickupPointName, *pickupPointAddress, *pickupPointContactDetails}
			}

		case "get-pickup-point":
			if *pickupPointName == "" {
				fmt.Println("не указано название пвз")
				continue
			}
			readCh <- []string{command, *pickupPointName}

		case "get-list-pickup-point":
			readCh <- []string{command, ""}

		case "help":
			fmt.Println("Commands:")
			fmt.Println("accept-ordrer: accept an order from a courier. " +
				"Arguments - ordid: string; custid: string; term-keeping: string (format: 2/1/2006); package-type: string (variants: packet, box, envelope); weight: int; price: int")
			fmt.Println("return-to-courier: return the order to the courier. Arguments - ordid: string")
			fmt.Println("give-order: issue an order to the client. Arguments - sliceid: slice of strings")
			fmt.Println("get-list-order: get a list of orders. Arguments - custid: string; limit: int; only-not-issued: bool")
			fmt.Println("refund-order: accept return from customer. Arguments - custid: string; ordid: string")
			fmt.Println("get-list-refund: get a list of returns. Arguments - page-number: int, default value: 1; size: int, default value: 5")
			fmt.Println("add-pickup-point: add data about pickup point to storage. Arguments - pickup-point-name: string; address: string; contact-details: string")
			fmt.Println("get-list-pickup-point: get a list of pickup points. No arguments")
			fmt.Println("get-pickup-point: get a pickup point. Arguments - pickup-point-name: string")

		default:
			fmt.Println("неизвестная команда")
		}
	}
}

func write(logch chan<- string, ch chan []string, pserv service.PickupPointService, oserv service.OrderService) {
	for {
		data := <-ch
		logch <- "Запись: начало обработки"
		command := data[0]
		args := data[1:]

		switch command {
		case "add-pickup-point":
			addPickupPoint(args, pserv)

		case "accept-order":
			acceptOrder(args, oserv)

		case "return-to-courier":
			returnToCourier(args, oserv)

		case "give-order":
			giveOrder(args, oserv)

		case "refund-order":
			refundOrder(args, oserv)

		}

		logch <- "Запись: конец обработки"
	}
}

func read(logch chan<- string, ch chan []string, pserv service.PickupPointService, oserv service.OrderService) {
	for {
		data := <-ch
		logch <- "Чтение: начало обработки"
		command := data[0]
		args := data[1:]

		switch command {
		case "get-list-pickup-point":
			getListPickupPoint(pserv)

		case "get-list-order":
			getListOrder(args, oserv)

		case "get-list-refund":
			getListRefund(args, oserv)

		case "get-pickup-point":
			getPickupPoint(args, pserv)
		}

		logch <- "Чтение: конец обработки"
	}
}

func logFunc(ch chan string) {
	for {
		message := <-ch
		fmt.Println(message)
	}
}

func addPickupPoint(args []string, serv service.PickupPointService) {
	err := serv.Write(args[0], args[1], args[2])
	if err != nil {
		fmt.Println("Ошибка: ", err)
	} else {
		fmt.Println("запись прошла успешна")
	}
}

func getListPickupPoint(serv service.PickupPointService) {
	pickupPoints := serv.Read()
	if len(pickupPoints) == 0 {
		fmt.Println("список пвз пуст")
	} else {
		for _, pickupPoint := range pickupPoints {
			rawBytes, err := json.Marshal(pickupPoint)
			if err != nil {
				fmt.Println("Ошибка:", err)
				break
			}
			fmt.Println(string(rawBytes))
		}
	}
}

func acceptOrder(args []string, serv service.OrderService) {
	date, err := time.Parse("2/1/2006", args[2])
	if err != nil {
		fmt.Println("Ошибка: ", err)
		return
	}

	weightTmp, err := strconv.ParseFloat(args[3], 32)
	if err != nil {
		fmt.Println("Ошибка: ", err)
		return
	}
	weight := float32(weightTmp)

	priceTmp, err := strconv.ParseFloat(args[4], 32)
	if err != nil {
		fmt.Println("Ошибка: ", err)
		return
	}
	price := float32(priceTmp)

	err = serv.AcceptOrderFromCourier(model.OrderInputDTO{
		OrderId:     args[0],
		CustomerId:  args[1],
		TermKeeping: date,
		Weight:      weight,
		Price:       price,
	}, args[5],
	)

	if err != nil {
		fmt.Println("Ошибка:", err)
		return
	}
	fmt.Println("заказ принят на склад")
}

func returnToCourier(args []string, serv service.OrderService) {
	err := serv.ReturnOrderToCourier(args[0])
	if err != nil {
		fmt.Println("Ошибка:", err)
		return
	}
	fmt.Println("заказ был возвращён курьеру")
}

func giveOrder(args []string, serv service.OrderService) {
	err := serv.GiveOrder(args)
	if err != nil {
		fmt.Println("Ошибка:", err)
		return
	}
	fmt.Println("заказы были успешно выданы")
}

func getListOrder(args []string, serv service.OrderService) {
	limit, err := strconv.Atoi(args[1])
	if err != nil {
		fmt.Println("Ошибка: ", err)
		return
	}

	onlyNotIssued, err := strconv.ParseBool(args[2])
	if err != nil {
		fmt.Println("Ошибка: ", err)
		return
	}

	orders, err := serv.GetOrders(args[0], limit, onlyNotIssued)
	if err != nil {
		fmt.Println("Ошибка:", err)
		return
	}

	if len(orders) == 0 {
		fmt.Println("список заказов пуст")
		return
	}

	for _, order := range orders {
		rawBytes, err := json.Marshal(order)
		if err != nil {
			fmt.Println("Ошибка:", err)
			return
		}
		fmt.Println(string(rawBytes))
	}
}

func refundOrder(args []string, serv service.OrderService) {
	err := serv.AcceptRefund(args[0], args[1])
	if err != nil {
		fmt.Println("Ошибка:", err)
		return
	}
	fmt.Println("заказ был успешно возвращён")
}

func getListRefund(args []string, serv service.OrderService) {
	pageNumber, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Println("Ошибка: ", err)
	}

	pageSize, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Println("Ошибка: ", err)
	}

	orders, err := serv.GetListRefund(pageNumber, pageSize)
	if err != nil {
		fmt.Println("Ошибка:", err)
		return
	}

	if len(orders) == 0 {
		fmt.Println(orders)
		return
	}

	for _, order := range orders {
		rawBytes, err := json.Marshal(order)
		if err != nil {
			fmt.Println("Ошибка:", err)
			return
		}
		fmt.Println(string(rawBytes))
	}
}

func getPickupPoint(args []string, serv service.PickupPointService) {
	pickupPoint, err := serv.GetPickupPoint(args[0])
	if err != nil {
		fmt.Println("Ошибка: ", err)
		return
	}
	rawBytes, err := json.Marshal(pickupPoint)
	if err != nil {
		fmt.Println("Ошибка:", err)
	}
	fmt.Println(string(rawBytes))
}

func httpMain() {
	if err := godotenv.Load("configs/config.env"); err != nil {
		log.Print("No .env file found")
		return
	}
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	database, err := db.NewDb(ctx)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer database.GetPool(ctx).Close()

	pickupPointsRepo := postgresql.NewPickupPoints(database)
	inMemoryCache := in_memory_cache.NewInMemoryCache()
	newRedis := redis2.NewRedis(&redis.Options{
		Addr:     os.Getenv("REDIS_HOST") + ":" + os.Getenv("REDIS_PORT"),
		Password: "",
		DB:       0,
	})
	businessService := domain.BusinessService{Repo: pickupPointsRepo, InMemoryCache: inMemoryCache, Redis: newRedis}
	implementation := server.Server{Service: &businessService}

	brokers := []string{os.Getenv("BROKER1"), os.Getenv("BROKER2"), os.Getenv("BROKER3")}

	kafkaProducer, err := kafka.NewProducer(brokers)
	if err != nil {
		fmt.Println("Ошибка:", err)
		return
	}
	kafkaSender := sender.NewKafkaSender(kafkaProducer, "logs")

	kafkaConsumer, err := kafka.NewConsumer(brokers)
	if err != nil {
		fmt.Println("Ошибка:", err)
		return
	}
	handlers := map[string]receiver.HandleFunc{
		"logs": func(message *sarama.ConsumerMessage) {
			pm := sender.LogMessage{}
			err = json.Unmarshal(message.Value, &pm)
			if err != nil {
				fmt.Println("Consumer error", err)
				return
			}
			fmt.Println("Received Key: ", string(message.Key), " Value: ", pm)
		},
	}
	logReceiver := receiver.NewReceiver(kafkaConsumer, handlers)
	logReceiver.Subscribe("logs")

	srv := &http.Server{
		Addr:    port,
		Handler: middleware.LogMiddleware(server.CreateRouter(implementation), *kafkaSender),
	}

	go func() {
		sig := <-sigCh
		fmt.Println()
		fmt.Println("Получен сигнал завершения программы:", sig)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err := srv.Shutdown(ctx)
		if err != nil {
			log.Println(err)
			return
		}

	}()

	log.Println("HTTP-сервер запущен на порту: ", port)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
		return
	}
}
