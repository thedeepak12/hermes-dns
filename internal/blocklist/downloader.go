package blocklist

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type Source struct {
	Name           string   `json:"name"`
	URL            string   `json:"url"`
	Format         string   `json:"format"`
	UpdateInterval Duration `json:"update_interval"`
	LocalPath      string   `json:"local_path"`
}

type Duration struct {
	time.Duration
}

func (d *Duration) UnmarshalJSON(b []byte) error {
	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}

	switch value := v.(type) {
	case float64:
		d.Duration = time.Duration(value)
	case string:
		var err error
		d.Duration, err = time.ParseDuration(value)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("invalid duration")
	}

	return nil
}

func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.Duration.String())
}

type Category struct {
	Enabled bool     `json:"enabled"`
	Sources []Source `json:"sources"`
}

type Registry struct {
	Blocklists map[string]Category `json:"blocklists"`
}

type Downloader struct {
	registryPath string
	registry     *Registry
	client       *http.Client
}

func NewDownloader(registryPath string) (*Downloader, error) {
	d := &Downloader{
		registryPath: registryPath,
		client:       &http.Client{Timeout: 30 * time.Second},
	}

	if err := d.loadRegistry(); err != nil {
		return nil, fmt.Errorf("failed to load registry: %w", err)
	}

	return d, nil
}

func (d *Downloader) loadRegistry() error {
	data, err := os.ReadFile(d.registryPath)
	if err != nil {
		return err
	}

	var registry Registry
	if err := json.Unmarshal(data, &registry); err != nil {
		return err
	}

	d.registry = &registry
	return nil
}

func (d *Downloader) DownloadAll() error {
	for category, catConfig := range d.registry.Blocklists {
		if !catConfig.Enabled {
			continue
		}

		for i := range catConfig.Sources {
			source := &catConfig.Sources[i]
			if d.shouldUpdate(source) {
				if err := d.downloadSource(source); err != nil {
					return fmt.Errorf("failed to download %s (%s): %w", source.Name, category, err)
				}
			}
		}
		d.registry.Blocklists[category] = catConfig
	}

	return nil
}

func (d *Downloader) shouldUpdate(source *Source) bool {
	info, err := os.Stat(source.LocalPath)
	if os.IsNotExist(err) {
		return true
	}
	if err != nil {
		return true
	}

	return time.Since(info.ModTime()) > source.UpdateInterval.Duration
}

func (d *Downloader) downloadSource(source *Source) error {
	resp, err := d.client.Get(source.URL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	if err := os.MkdirAll(filepath.Dir(source.LocalPath), 0755); err != nil {
		return err
	}

	tmpPath := source.LocalPath + ".tmp"
	out, err := os.Create(tmpPath)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, resp.Body); err != nil {
		return err
	}

	return os.Rename(tmpPath, source.LocalPath)
}

func (d *Downloader) saveRegistry() error {
	data, err := json.MarshalIndent(d.registry, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(d.registryPath, data, 0644)
}

func (d *Downloader) GetEnabledSources() map[string][]string {
	res := make(map[string][]string)
	for category, catConfig := range d.registry.Blocklists {
		if !catConfig.Enabled {
			continue
		}
		for _, source := range catConfig.Sources {
			res[category] = append(res[category], source.LocalPath)
		}
	}

	return res
}
