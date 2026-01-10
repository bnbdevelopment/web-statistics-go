package geolocation

import (
	"errors"
	"log"
	"net"
	"sync"

	"github.com/oschwald/geoip2-golang"
)

// GeoService provides thread-safe IP geolocation lookups
type GeoService struct {
	reader *geoip2.Reader
	mu     sync.RWMutex // For future reload capability
}

// GeoData represents the extracted geolocation information
type GeoData struct {
	CountryCode string
	CountryName string
	City        string
	Region      string
	Latitude    float64
	Longitude   float64
}

var (
	service  *GeoService
	once     sync.Once
	localIps = map[string]struct{}{ // Use a map for efficient lookups
		"127.0.0.1": {},
		"::1":       {},
	}
)

// InitGeoService initializes the global GeoIP reader (call once at startup)
func InitGeoService(dbPath string) error {
	var initErr error
	once.Do(func() {
		reader, err := geoip2.Open(dbPath)
		if err != nil {
			initErr = err
			return
		}
		service = &GeoService{reader: reader}
		log.Println("GeoIP database loaded successfully from:", dbPath)
	})
	return initErr
}

// ErrServiceNotInitialized is returned when the geolocation service has not been initialized.
var ErrServiceNotInitialized = errors.New("geolocation service not initialized")

// ErrInvalidIP is returned for invalid IP addresses.
var ErrInvalidIP = errors.New("invalid IP address")

// Lookup performs IP geolocation lookup
func Lookup(ipStr string) (*GeoData, error) {
	if _, isLocal := localIps[ipStr]; isLocal {
		return &GeoData{
			CountryCode: "00",
			CountryName: "Localhost",
			Latitude:    0,
			Longitude:   0,
		}, nil
	}

	if service == nil || service.reader == nil {
		return nil, ErrServiceNotInitialized
	}

	service.mu.RLock()
	defer service.mu.RUnlock()

	ip := net.ParseIP(ipStr)
	if ip == nil {
		return nil, ErrInvalidIP
	}

	record, err := service.reader.City(ip)
	if err != nil {
		// Log the error but also return it, wrapped for context.
		log.Printf("GeoIP lookup failed for %s: %v", ipStr, err)
		return nil, err
	}

	geoData := &GeoData{
		CountryCode: record.Country.IsoCode,
		CountryName: record.Country.Names["en"],
		Latitude:    record.Location.Latitude,
		Longitude:   record.Location.Longitude,
	}

	if len(record.City.Names) > 0 {
		geoData.City = record.City.Names["en"]
	}

	if len(record.Subdivisions) > 0 {
		geoData.Region = record.Subdivisions[0].Names["en"]
	}

	return geoData, nil
}

// Close releases the GeoIP database resources
func Close() error {
	if service != nil && service.reader != nil {
		return service.reader.Close()
	}
	return nil
}
