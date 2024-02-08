/*
favirecon - Use favicon.ico to improve your target recon phase. Quickly detect technologies, WAF, exposed panels, known services.

This repository is under MIT License https://github.com/edoardottt/favirecon/blob/main/LICENSE
*/

package favirecon

import "go.uber.org/ratelimit"

func rateLimiter(r *Runner) ratelimit.Limiter {
	var ratelimiter ratelimit.Limiter
	if r.Options.RateLimit > 0 {
		ratelimiter = ratelimit.New(r.Options.RateLimit)
	} else {
		ratelimiter = ratelimit.NewUnlimited()
	}

	return ratelimiter
}
