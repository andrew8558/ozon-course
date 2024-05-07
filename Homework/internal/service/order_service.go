package service

import (
	"Homework/internal/model"
	"errors"
	"time"
)

type orderStorage interface {
	Save(order model.Order) error
	Delete(orderId string) error
	List() ([]model.Order, error)
}

type packager interface {
	Pack(order model.OrderInputDTO) (model.OrderInputDTO, error)
}

type OrderService struct {
	o orderStorage
}

func NewOrderService(s orderStorage) OrderService {
	return OrderService{o: s}
}

func (s OrderService) AcceptOrderFromCourier(orderInput model.OrderInputDTO, packageType string) error {
	orders, err := s.o.List()
	if err != nil {
		return err
	}

	packer, err := s.makePacker(packageType)
	if err != nil {
		return err
	}

	orderInput, err = packer.Pack(orderInput)
	if err != nil {
		return err
	}

	curTime := time.Now()
	if curTime.After(orderInput.TermKeeping) {
		return errors.New("срок хранения в прошлом")
	}

	if s.getOrderIndex(orders, orderInput.OrderId) != -1 {
		return errors.New("заказ существует")
	}

	return s.o.Save(model.Order{
		OrderID:     orderInput.OrderId,
		CustomerId:  orderInput.CustomerId,
		TermKeeping: orderInput.TermKeeping,
		Status:      model.Accepted,
		Weight:      orderInput.Weight,
		Price:       orderInput.Price,
	})
}

func (s OrderService) makePacker(packageType string) (packager, error) {
	switch packageType {
	case "packet":
		return &model.BatchPacker{}, nil
	case "box":
		return &model.BoxPacker{}, nil
	case "envelope":
		return &model.EnvelopePacker{}, nil
	default:
		return nil, errors.New("данного вида упаковки не существует")
	}
}

func (s OrderService) getOrderIndex(orders []model.Order, orderId string) int {
	for index, order := range orders {
		if order.OrderID == orderId {
			return index
		}
	}
	return -1
}

func (s OrderService) ReturnOrderToCourier(orderId string) error {
	orders, err := s.o.List()
	if err != nil {
		return nil
	}

	index := s.getOrderIndex(orders, orderId)
	if index == -1 {
		return errors.New("заказ не найден")
	}
	order := orders[index]

	if order.Status == model.Issued {
		return errors.New("заказ был выдан клиенту")
	}

	if !order.TermKeeping.Before(time.Now()) {
		return errors.New("срок хранения заказа ещё не истёк")
	}

	return s.o.Delete(orderId)
}

func (s OrderService) GiveOrder(listOrderId []string) error {
	orders, err := s.o.List()
	if err != nil {
		return err
	}

	listOrder := make([]model.Order, 0)
	var customerId = ""
	for _, orderId := range listOrderId {
		index := s.getOrderIndex(orders, orderId)
		if index == -1 {
			return errors.New("заказы не найдены")
		}

		if customerId == "" {
			customerId = orders[index].CustomerId
		} else if customerId != orders[index].CustomerId {
			return errors.New("id заказов принадлежат не одному человеку")
		}
		listOrder = append(listOrder, orders[index])
	}

	for _, order := range listOrder {
		if order.Status == model.Refunded {
			return errors.New("товар был возвращён")
		}
		if order.Status == model.Issued {
			return errors.New("товар был выдан")
		}
		if order.TermKeeping.Before(time.Now()) {
			return errors.New("срок хранения заказа истёк")
		}
		order.Status = model.Issued
		order.DateIssue = time.Now()
		err = s.o.Save(order)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s OrderService) GetOrders(customerId string, limit int, onlyNotIssued bool) ([]model.OrderOutputDTO, error) {
	orders, err := s.o.List()
	if err != nil {
		return nil, err
	}

	listOrderOutputDTO := make([]model.OrderOutputDTO, 0)
	for _, order := range orders {
		if order.CustomerId == customerId && (!onlyNotIssued || !(order.Status == model.Issued)) {
			listOrderOutputDTO = append(listOrderOutputDTO, model.OrderOutputDTO{
				OrderId:     order.OrderID,
				CustomerId:  order.CustomerId,
				TermKeeping: order.TermKeeping,
				Status:      order.Status,
				Weight:      order.Weight,
				Price:       order.Price,
			})
		}
	}

	if limit == 0 || limit >= len(listOrderOutputDTO) {
		return listOrderOutputDTO, nil
	} else {
		return listOrderOutputDTO[len(listOrderOutputDTO)-limit:], nil
	}
}

func (s OrderService) AcceptRefund(customerId string, orderId string) error {
	orders, err := s.o.List()
	if err != nil {
		return err
	}

	index := s.getOrderIndex(orders, orderId)
	if index == -1 {
		return errors.New("заказ не найден")
	}

	if orders[index].CustomerId != customerId {
		return errors.New("заказ не принадлежит данному клиенту")
	}

	if orders[index].Status != model.Issued {
		return errors.New("заказ ещё не был выдан")
	}

	if orders[index].DateIssue.AddDate(0, 0, 2).Before(time.Now()) {
		return errors.New("период возврата истёк")
	}

	orders[index].Status = model.Refunded
	return s.o.Save(orders[index])
}

func (s OrderService) GetListRefund(pageNumber int, pageSize int) ([]model.OrderOutputDTO, error) {
	orders, err := s.o.List()
	if err != nil {
		return nil, err
	}

	listOrderOutputDTO := make([]model.OrderOutputDTO, 0)
	for _, order := range orders {
		if order.Status == model.Refunded {
			listOrderOutputDTO = append(listOrderOutputDTO, model.OrderOutputDTO{
				OrderId:     order.OrderID,
				CustomerId:  order.CustomerId,
				TermKeeping: order.TermKeeping,
				Status:      order.Status,
				Weight:      order.Weight,
				Price:       order.Price,
			})
		}
	}
	if len(listOrderOutputDTO) == 0 {
		return listOrderOutputDTO, nil
	}

	rightIndex := pageNumber * pageSize
	leftIndex := rightIndex - pageSize
	if rightIndex > len(listOrderOutputDTO) {
		rightIndex = len(listOrderOutputDTO)
	}

	if leftIndex >= rightIndex {
		return make([]model.OrderOutputDTO, 0), nil
	}

	return listOrderOutputDTO[leftIndex:rightIndex], nil

}
