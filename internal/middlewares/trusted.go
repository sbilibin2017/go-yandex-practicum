package middlewares

import (
	"net"
	"net/http"
	"strings"
)

// NewTrustedSubnetMiddleware возвращает middleware, который проверяет,
// содержит ли заголовок X-Real-IP входящего запроса IP-адрес из указанной доверенной подсети.
//
// Если заголовок отсутствует, имеет некорректный формат или IP-адрес не входит в доверенную подсеть,
// middleware отвечает статусом HTTP 403 Forbidden и не передаёт запрос следующему обработчику.
//
// Доверенная подсеть должна быть указана в CIDR-нотации (например, "192.168.1.0/24").
// Если CIDR не может быть разобран при инициализации, middleware будет отклонять все входящие запросы.
func TrustedSubnetMiddleware(trustedSubnet string) func(http.Handler) http.Handler {
	_, subnet, err := net.ParseCIDR(trustedSubnet)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Отклонить все запросы, если подсеть некорректна
			if err != nil || subnet == nil {
				w.WriteHeader(http.StatusForbidden)
				return
			}

			// Получить и проверить заголовок X-Real-IP
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

			// IP валиден и доверенный, передать запрос дальше
			next.ServeHTTP(w, r)
		})
	}
}
