package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/shopspring/decimal"
	"stalll-hub-pos/backend/config"
	"stalll-hub-pos/backend/internal/dto"
)

type AmapService struct {
	cfg    *config.AmapConfig
	client *http.Client
}

func NewAmapService() *AmapService {
	return &AmapService{
		cfg:    &config.AppConfig.Amap,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (s *AmapService) PlanRoute(originLng, originLat, destLng, destLat float64) (*dto.RoutePlanResponse, error) {
	if s.cfg.Key == "" {
		return nil, fmt.Errorf("amap key not configured")
	}

	origin := fmt.Sprintf("%.6f,%.6f", originLng, originLat)
	destination := fmt.Sprintf("%.6f,%.6f", destLng, destLat)

	params := url.Values{}
	params.Set("key", s.cfg.Key)
	params.Set("origin", origin)
	params.Set("destination", destination)
	params.Set("extensions", "base")
	params.Set("strategy", "0")

	apiURL := fmt.Sprintf("%s/direction/driving?%s", s.cfg.BaseURL, params.Encode())

	resp, err := s.client.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("amap route planning request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read amap response failed: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse amap response failed: %w", err)
	}

	status, _ := result["status"].(string)
	if status != "1" {
		info, _ := result["info"].(string)
		return nil, fmt.Errorf("amap api error: %s", info)
	}

	routeData, _ := result["route"].(map[string]interface{})
	paths, _ := routeData["paths"].([]interface{})
	if len(paths) == 0 {
		return nil, fmt.Errorf("no route found")
	}

	firstPath, _ := paths[0].(map[string]interface{})
	distanceStr, _ := firstPath["distance"].(string)
	durationStr, _ := firstPath["duration"].(string)

	distance, _ := strconv.ParseFloat(distanceStr, 64)
	duration, _ := strconv.Atoi(durationStr)

	distanceKm := distance / 1000.0
	fee := calculateDeliveryFee(distanceKm)

	routeJSON, _ := json.Marshal(firstPath)

	return &dto.RoutePlanResponse{
		Distance: distanceKm,
		Duration: duration / 60,
		Route:    string(routeJSON),
		Fee:      fee,
	}, nil
}

func (s *AmapService) Geocode(address string, city string) (*dto.GeocodeResponse, error) {
	if s.cfg.Key == "" {
		return nil, fmt.Errorf("amap key not configured")
	}

	params := url.Values{}
	params.Set("key", s.cfg.Key)
	params.Set("address", address)
	if city != "" {
		params.Set("city", city)
	}

	apiURL := fmt.Sprintf("%s/geocode/geo?%s", s.cfg.BaseURL, params.Encode())

	resp, err := s.client.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("amap geocode request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read amap response failed: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse amap response failed: %w", err)
	}

	status, _ := result["status"].(string)
	if status != "1" {
		info, _ := result["info"].(string)
		return nil, fmt.Errorf("amap api error: %s", info)
	}

	geocodes, _ := result["geocodes"].([]interface{})
	if len(geocodes) == 0 {
		return nil, fmt.Errorf("no geocode result found")
	}

	geocode, _ := geocodes[0].(map[string]interface{})
	location, _ := geocode["location"].(string)
	formatted, _ := geocode["formatted_address"].(string)

	var lng, lat float64
	fmt.Sscanf(location, "%f,%f", &lng, &lat)

	return &dto.GeocodeResponse{
		Lng:       lng,
		Lat:       lat,
		Formatted: formatted,
	}, nil
}

func (s *AmapService) Regeocode(lng, lat float64) (string, error) {
	if s.cfg.Key == "" {
		return "", fmt.Errorf("amap key not configured")
	}

	location := fmt.Sprintf("%.6f,%.6f", lng, lat)
	params := url.Values{}
	params.Set("key", s.cfg.Key)
	params.Set("location", location)
	params.Set("extensions", "base")

	apiURL := fmt.Sprintf("%s/geocode/regeo?%s", s.cfg.BaseURL, params.Encode())

	resp, err := s.client.Get(apiURL)
	if err != nil {
		return "", fmt.Errorf("amap regeocode request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read amap response failed: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("parse amap response failed: %w", err)
	}

	status, _ := result["status"].(string)
	if status != "1" {
		info, _ := result["info"].(string)
		return "", fmt.Errorf("amap api error: %s", info)
	}

	regeocode, _ := result["regeocode"].(map[string]interface{})
	formatted, _ := regeocode["formatted_address"].(string)
	return formatted, nil
}

func (s *AmapService) GetRiderRoute(riderLng, riderLat, destLng, destLat float64) (*dto.RoutePlanResponse, error) {
	return s.PlanRoute(riderLng, riderLat, destLng, destLat)
}

func calculateDeliveryFee(distanceKm float64) decimal.Decimal {
	baseFee := 5.0
	perKmFee := 1.5
	if distanceKm <= 3 {
		return decimal.NewFromFloat(baseFee)
	}
	extra := (distanceKm - 3) * perKmFee
	return decimal.NewFromFloat(baseFee + extra)
}
