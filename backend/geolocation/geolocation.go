package geolocation

import (
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
	service *GeoService
	once    sync.Once
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

// Lookup performs IP geolocation lookup
func Lookup(ipStr string) (*GeoData, error) {
	if service == nil {
		return nil, nil // Graceful degradation if service not initialized
	}

	service.mu.RLock()
	defer service.mu.RUnlock()

	ip := net.ParseIP(ipStr)
	if ip == nil {
		return nil, nil // Invalid IP - graceful degradation
	}

	record, err := service.reader.City(ip)
	if err != nil {
		// Log but don't fail - graceful degradation
		log.Printf("GeoIP lookup failed for %s: %v", ipStr, err)
		return nil, nil
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
