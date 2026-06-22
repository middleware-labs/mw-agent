package agent

import (
	"bytes"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/url"
	"runtime"
	"runtime/pprof"
	"time"

	"github.com/k0kubun/pp"
	"go.uber.org/zap"
)

var validProfileTypes = map[string]bool{
	"cpu":       true,
	"heap":      true,
	"goroutine": true,
	"allocs":    true,
	"block":     true,
	"mutex":     true,
}

func (c *HostAgent) collectAndUploadProfiles(req *profilingRequest) {
	if !c.profilingInProgress.CompareAndSwap(false, true) {
		c.logger.Warn("profiling already in progress, skipping request",
			zap.String("request_id", req.RequestID))
		return
	}
	defer c.profilingInProgress.Store(false)

	pp.Println("[profiling] starting collection for request:", req.RequestID)
	pp.Println("[profiling] requested types:", req.Types)

	c.logger.Info("starting remote profiling",
		zap.String("request_id", req.RequestID),
		zap.Strings("types", req.Types))

	cpuDuration := time.Duration(req.CPUDurationSeconds) * time.Second
	if cpuDuration <= 0 || cpuDuration > 120*time.Second {
		cpuDuration = 30 * time.Second
	}
	pp.Println("[profiling] cpu duration:", cpuDuration)

	for _, profileType := range req.Types {
		if !validProfileTypes[profileType] {
			pp.Println("[profiling] unknown profile type, skipping:", profileType)
			c.logger.Warn("unknown profile type, skipping",
				zap.String("type", profileType),
				zap.String("request_id", req.RequestID))
			continue
		}

		pp.Println("[profiling] capturing profile:", profileType)
		data, err := captureProfile(profileType, cpuDuration)
		if err != nil {
			pp.Println("[profiling] capture failed:", profileType, err)
			c.logger.Error("failed to capture profile",
				zap.String("type", profileType),
				zap.String("request_id", req.RequestID),
				zap.Error(err))
			continue
		}
		pp.Println("[profiling] captured:", profileType, "size_bytes:", len(data))

		pp.Println("[profiling] uploading:", profileType)
		if err := c.uploadProfile(req.RequestID, profileType, data); err != nil {
			pp.Println("[profiling] upload failed:", profileType, err)
			c.logger.Error("failed to upload profile",
				zap.String("type", profileType),
				zap.String("request_id", req.RequestID),
				zap.Error(err))
		} else {
			pp.Println("[profiling] upload success:", profileType, "size_bytes:", len(data))
			c.logger.Info("profile uploaded successfully",
				zap.String("type", profileType),
				zap.String("request_id", req.RequestID),
				zap.Int("size_bytes", len(data)))
		}
	}

	pp.Println("[profiling] collection complete for request:", req.RequestID)
}

func captureProfile(profileType string, cpuDuration time.Duration) ([]byte, error) {
	var buf bytes.Buffer

	switch profileType {
	case "cpu":
		if err := pprof.StartCPUProfile(&buf); err != nil {
			return nil, fmt.Errorf("start CPU profile: %w", err)
		}
		time.Sleep(cpuDuration)
		pprof.StopCPUProfile()

	case "heap":
		runtime.GC()
		if err := pprof.WriteHeapProfile(&buf); err != nil {
			return nil, fmt.Errorf("write heap profile: %w", err)
		}

	default:
		p := pprof.Lookup(profileType)
		if p == nil {
			return nil, fmt.Errorf("unknown profile: %s", profileType)
		}
		if err := p.WriteTo(&buf, 0); err != nil {
			return nil, fmt.Errorf("write %s profile: %w", profileType, err)
		}
	}

	return buf.Bytes(), nil
}

func (c *HostAgent) uploadProfile(requestID, profileType string, data []byte) error {
	hostname := GetHostnameForPlatform(c.InfraPlatform)

	u, err := url.Parse(c.APIURLForConfigCheck)
	if err != nil {
		return err
	}
	baseURL := u.JoinPath(apiPathForProfilingUpload, c.APIKey)

	pp.Println("[profiling] upload URL:", baseURL.String())
	pp.Println("[profiling] upload host_id:", hostname)

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	_ = writer.WriteField("request_id", requestID)
	_ = writer.WriteField("profile_type", profileType)
	_ = writer.WriteField("host_id", hostname)

	filename := fmt.Sprintf("%s_%s.prof", requestID, profileType)
	part, err := writer.CreateFormFile("profile", filename)
	if err != nil {
		return fmt.Errorf("create form file: %w", err)
	}
	if _, err := part.Write(data); err != nil {
		return fmt.Errorf("write profile data: %w", err)
	}
	writer.Close()

	req, err := http.NewRequest(http.MethodPost, baseURL.String(), &body)
	if err != nil {
		return fmt.Errorf("create upload request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("upload request failed: %w", err)
	}
	defer resp.Body.Close()

	pp.Println("[profiling] upload response status:", resp.StatusCode)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("upload returned status %d", resp.StatusCode)
	}

	return nil
}
