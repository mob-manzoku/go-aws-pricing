package goawspricing

import (
	"io/ioutil"
	"strconv"

	"github.com/bitly/go-simplejson"
	"github.com/mob-manzoku/go-jsonp"
	"gopkg.in/yaml.v2"
)

type InstanceTypes map[string]*Size

type Size struct {
	Size              *string  `json:"size"`
	VCPU              *int     `json:"vCPU"`
	ECU               *string  `json:"ECU"`
	MemoryGiB         *float64 `json:"memoryGiB"`
	StorageGB         *string  `json:"storageGB"`
	System            *string  `json:"system"`
	PiopsOptimized    *bool    `json:"piopsOptimized"`
	PriceHour         *float64 `json:"price_hour"`
	PriceDay          *float64 `json:"price_day"`
	PriceMonth        *float64 `json:"price_month"`
	Network           *string  `json:"network"`
	GP2StoragePrice   *float64 `json:"gp2_storage_price"`
	PIOPSStoragePrice *float64 `json:"piops_storage_price"`
	PIOPSIOPrice      *float64 `json:"piops_io_price"`
}

var awsPricingURLs = map[string]string{
	"ec2":         "http://a0.awsstatic.com/pricing/1/ec2/linux-od.min.js",
	"rds":         "http://a0.awsstatic.com/pricing/1/rds/mysql/pricing-standard-deployments.min.js",
	"elasticache": "http://a0.awsstatic.com/pricing/1/elasticache/pricing-standard-deployments-elasticache.min.js",
	"rds_gp2":     "https://a0.awsstatic.com/pricing/1/rds/mysql/pricing-gp2-standard-deploy.min.js",
	"rds_piops":   "https://a0.awsstatic.com/pricing/1/rds/mysql/pricing-piops-standard-deploy.min.js",
	"ec2_ebs":     "https://a0.awsstatic.com/pricing/1/ebs/pricing-ebs.min.js",
}

func priceMultiplication(hourPrice float64) (float64, float64) {
	hourDay := float64(24)
	daysMonth := float64(30)
	significant := float64(100)

	dayPrice := float64(int(hourPrice*hourDay*significant)) / significant
	monthPrice := float64(int(hourPrice*hourDay*daysMonth*significant)) / significant

	return dayPrice, monthPrice
}

func GetEC2Pricing(region string) InstanceTypes {

	service := "ec2"
	ret := InstanceTypes{}

	str, _ := gojsonp.GetJSONFromURL(awsPricingURLs[service])

	raw, err := simplejson.NewJson([]byte(str))
	if err != nil {
		panic(err)
	}

	regions := raw.Get("config").Get("regions")

	for i := range regions.MustArray() {

		if regions.GetIndex(i).Get("region").MustString() != region {
			continue
		}

		types := regions.GetIndex(i).Get("instanceTypes")

		for j := range types.MustArray() {

			sizes := types.GetIndex(j).Get("sizes")
			for k := range sizes.MustArray() {
				s := sizes.GetIndex(k)

				msize := s.Get("size").MustString()
				mvCPU, _ := strconv.Atoi(s.Get("vCPU").MustString())
				mECU := s.Get("ECU").MustString()
				mmem, _ := strconv.ParseFloat(s.Get("memoryGiB").MustString(), 64)
				msto := s.Get("storageGB").MustString()
				msys := s.Get("valueColumns").GetIndex(0).Get("name").MustString()
				mprice, _ := strconv.ParseFloat(s.Get("valueColumns").GetIndex(0).Get("prices").Get("USD").MustString(), 64)
				mpriced, mpricem := priceMultiplication(mprice)

				obj := &Size{
					Size:       &msize,
					VCPU:       &mvCPU,
					ECU:        &mECU,
					MemoryGiB:  &mmem,
					StorageGB:  &msto,
					System:     &msys,
					PriceHour:  &mprice,
					PriceDay:   &mpriced,
					PriceMonth: &mpricem,
				}

				ret[*obj.Size] = obj

			}
		}

	}
	return ret

}

func GetElasticachePricing(region string) InstanceTypes {

	service := "elasticache"
	specpath := "spec/elasticache.yml"
	ret := InstanceTypes{}

	str, _ := gojsonp.GetJSONFromURL(awsPricingURLs[service])

	raw, err := simplejson.NewJson([]byte(str))
	if err != nil {
		panic(err)
	}

	regions := raw.Get("config").Get("regions")

	for i := range regions.MustArray() {

		if regions.GetIndex(i).Get("region").MustString() != region {
			continue
		}

		types := regions.GetIndex(i).Get("types")

		for j := range types.MustArray() {

			tiers := types.GetIndex(j).Get("tiers")

			for k := range tiers.MustArray() {

				msize := tiers.GetIndex(k).Get("name").MustString()
				mprice, _ := strconv.ParseFloat(tiers.GetIndex(k).Get("prices").Get("USD").MustString(), 64)
				mpriced, mpricem := priceMultiplication(mprice)
				msys := "redis"

				obj := &Size{
					Size:       &msize,
					PriceHour:  &mprice,
					PriceDay:   &mpriced,
					PriceMonth: &mpricem,
					System:     &msys,
				}

				ret[*obj.Size] = obj
			}

		}
	}

	buf, err := ioutil.ReadFile(specpath)
	if err != nil {
		return ret
	}

	m := make(map[interface{}]interface{})
	err = yaml.Unmarshal(buf, &m)

	for _, v := range m[service].([]interface{}) {

		msize := v.(map[interface{}]interface{})["size"].(string)
		mvCPU := v.(map[interface{}]interface{})["vCPU"].(int)
		mmem := v.(map[interface{}]interface{})["memoryGiB"].(float64)
		mnet := v.(map[interface{}]interface{})["network"].(string)

		if _, ok := ret[msize]; ok {
			ret[msize].VCPU = &mvCPU
			ret[msize].MemoryGiB = &mmem
			ret[msize].Network = &mnet
		}

	}

	return ret
}

func GetRDSPricing(region string) InstanceTypes {

	service := "rds"
	specpath := "spec/rds.yml"
	ret := InstanceTypes{}
	str, _ := gojsonp.GetJSONFromURL(awsPricingURLs[service])

	raw, err := simplejson.NewJson([]byte(str))
	if err != nil {
		panic(err)
	}

	regions := raw.Get("config").Get("regions")

	for i := range regions.MustArray() {

		if regions.GetIndex(i).Get("region").MustString() != region {
			continue
		}

		types := regions.GetIndex(i).Get("types")

		for j := range types.MustArray() {

			tiers := types.GetIndex(j).Get("tiers")

			for k := range tiers.MustArray() {

				msize := tiers.GetIndex(k).Get("name").MustString()
				mprice, _ := strconv.ParseFloat(tiers.GetIndex(k).Get("prices").Get("USD").MustString(), 64)
				mpriced, mpricem := priceMultiplication(mprice)
				msys := "mysql"

				obj := &Size{
					Size:       &msize,
					PriceHour:  &mprice,
					PriceDay:   &mpriced,
					PriceMonth: &mpricem,
					System:     &msys,
				}

				ret[*obj.Size] = obj
			}

		}
	}

	buf, err := ioutil.ReadFile(specpath)
	if err != nil {
		return ret
	}

	m := make(map[interface{}]interface{})
	err = yaml.Unmarshal(buf, &m)

	for _, v := range m[service].([]interface{}) {

		msize := v.(map[interface{}]interface{})["size"].(string)
		mvCPU := v.(map[interface{}]interface{})["vCPU"].(int)
		mpiop := v.(map[interface{}]interface{})["piopsOptimized"].(bool)
		mmem := v.(map[interface{}]interface{})["memoryGiB"].(float64)
		mnet := v.(map[interface{}]interface{})["network"].(string)

		if _, ok := ret[msize]; ok {
			ret[msize].VCPU = &mvCPU
			ret[msize].MemoryGiB = &mmem
			ret[msize].PiopsOptimized = &mpiop
			ret[msize].Network = &mnet
		}

	}

	return ret
}

// GetEC2GP2Pricing is getting pricing for EC2 GP2 storage. The return value is monthly fee per GB.
func GetEC2GP2Pricing(region string) float64 {
	GP2 := "Amazon EBS General Purpose (SSD) volumes"

	str, _ := gojsonp.GetJSONFromURL(awsPricingURLs["ec2_ebs"])

	raw, _ := simplejson.NewJson([]byte(str))
	regions := raw.Get("config").Get("regions")
	ret := 0.0
	for i := range regions.MustArray() {

		if regions.GetIndex(i).Get("region").MustString() != region {
			continue
		}

		types := regions.GetIndex(i).Get("types")
		for t := range types.MustArray() {
			if types.GetIndex(t).Get("name").MustString() != GP2 {
				continue
			}

			ret, _ = strconv.ParseFloat(types.GetIndex(t).Get("values").GetIndex(0).Get("prices").Get("USD").MustString(), 64)
		}
	}
	return ret
}

// GetRDSGP2Pricing is getting pricing for RDS(Mysql) GP2 storage. The return value is monthly fee per GB.
func GetRDSGP2Pricing(region string) float64 {
	str, _ := gojsonp.GetJSONFromURL(awsPricingURLs["rds_gp2"])
	raw, _ := simplejson.NewJson([]byte(str))

	regions := raw.Get("config").Get("regions")
	ret := 0.0

	for i := range regions.MustArray() {

		if regions.GetIndex(i).Get("region").MustString() != region {
			continue
		}

		ret, _ = strconv.ParseFloat(regions.GetIndex(i).Get("rates").GetIndex(0).Get("prices").Get("USD").MustString(), 64)
	}

	return ret
}

// GetRDSPIOPSPricing is getting pricing for RDS(Mysql) PIOPS storage. The return values are monthly fee per GB.
func GetRDSPIOPSPricing(region string) (storage float64, io float64) {
	str, _ := gojsonp.GetJSONFromURL(awsPricingURLs["rds_piops"])
	raw, _ := simplejson.NewJson([]byte(str))

	regions := raw.Get("config").Get("regions")

	for i := range regions.MustArray() {

		if regions.GetIndex(i).Get("region").MustString() != region {
			continue
		}

		for j := range regions.GetIndex(i).Get("rates").MustArray() {

			if regions.GetIndex(i).Get("rates").GetIndex(j).Get("type").MustString() == "storageRate" {
				storage, _ = strconv.ParseFloat(regions.GetIndex(i).Get("rates").GetIndex(j).Get("prices").Get("USD").MustString(), 64)
			}
			if regions.GetIndex(i).Get("rates").GetIndex(j).Get("type").MustString() == "piopsRate" {
				io, _ = strconv.ParseFloat(regions.GetIndex(i).Get("rates").GetIndex(j).Get("prices").Get("USD").MustString(), 64)
			}
		}
	}

	return storage, io
}
