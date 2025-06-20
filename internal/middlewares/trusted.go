package middlewares

import (
	"net"
	"net/http"
	"strings"
)

// TrustedSubnetMiddleware возвращает middleware, который проверяет,
// содержит ли заголовок X-Real-IP входящего запроса IP-адрес из указанной доверенной подсети.
//
// Если trustedSubnet равен nil — проверка пропускается и все запросы проходят.
//
// Если заголовок отсутствует, имеет некорректный формат или IP-адрес не входит в доверенную подсеть,
// middleware отвечает статусом HTTP 403 Forbidden и не передаёт запрос следующему обработчику.
//
// Доверенная подсеть должна быть указана в CIDR-нотации (например, "192.168.1.0/24").
func TrustedSubnetMiddleware(trustedSubnet string) func(http.Handler) http.Handler {
	var subnet *net.IPNet
	if trustedSubnet != "" {
		_, parsedSubnet, err := net.ParseCIDR(trustedSubnet)
		if err == nil {
			subnet = parsedSubnet
		}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Если trustedSubnet не задан или CIDR некорректен — пропускаем проверку
			if subnet == nil {
				next.ServeHTTP(w, r)
				return
			}

			ipStr := strings.TrimSpace(r.Header.Get("X-Real-IP"))
			if ipStr == "" {
				w.WriteHeader(http.StatusForbidden)
				return
			}

			ip := net.ParseIP(ipStr)
			if ip == nil || !subnet.Contains(ip) {
				w.WriteHeader(http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
