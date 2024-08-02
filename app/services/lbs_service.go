package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	apperror "github.com/lixiaofei123/daily/app/errors"
	"github.com/lixiaofei123/daily/configs"
)

type Address struct {
	IP      string    `json:"ip"`
	Point   []float64 `json:"point"`
	Address []string  `json:"address"`
}

type LBSService interface {
	initConfig(config map[string]string)

	GetAddressByIp(ip string) (*Address, error)

	IsSupport() bool
}

var lbsServices map[string]LBSService

func init() {
	lbsServices = map[string]LBSService{}

	lbsServices["baidu"] = &BDLBSService{}
}

func NewLBSService() LBSService {

	lbsConfig := configs.GlobalConfig.LBSConfig
	name := lbsConfig.Name
	lbsService, ok := lbsServices[strings.ToLower(name)]
	if ok {
		lbsService.initConfig(lbsConfig.Config)
		return lbsService
	}

	return &EmptyLBSService{}
}

type BDLBSService struct {
	ak string
}

type BDPoint struct {
	X string `json:"x"`
	Y string `json:"y"`
}

type BDIPAddressDetail struct {
	Province string `json:"province"`
	City     string `json:"city"`
	Street   string `json:"street"`
}

type BDIPAddressContent struct {
	AddressDetail BDIPAddressDetail `json:"address_detail"`
	Point         BDPoint           `json:"point"`
}

type BDIPAddressResp struct {
	Status  int                `json:"status"`
	Content BDIPAddressContent `json:"content"`
}

func (bs *BDLBSService) initConfig(config map[string]string) {
	bs.ak = config["ak"]
}

func (bs *BDLBSService) GetAddressByIp(ip string) (*Address, error) {

	resp, err := http.DefaultClient.Get(
		fmt.Sprintf("https://api.map.baidu.com/location/ip?ip=%s&coor=bd09ll&ak=%s", ip, bs.ak))
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		resp := new(BDIPAddressResp)
		err = json.Unmarshal(data, &resp)
		if err != nil {
			return nil, err
		}

		if resp.Status == 0 {
			point := resp.Content.Point
			point_x, _ := strconv.ParseFloat(point.X, 64)
			point_y, _ := strconv.ParseFloat(point.Y, 64)
			baaddress := resp.Content.AddressDetail
			return &Address{
				IP:      ip,
				Address: []string{baaddress.Province, baaddress.City, baaddress.Street},
				Point:   []float64{point_x, point_y},
			}, nil
		}

	}

	return nil, apperror.ErrLocationError
}

func (ls *BDLBSService) IsSupport() bool {
	return true
}

type EmptyLBSService struct {
}

func (*EmptyLBSService) initConfig(config map[string]string) {

}

func (*EmptyLBSService) GetAddressByIp(ip string) (*Address, error) {
	return nil, errors.ErrUnsupported
}

func (*EmptyLBSService) IsSupport() bool {
	return false
}
