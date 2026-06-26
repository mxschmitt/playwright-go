package playwright

import (
	"fmt"
	"strings"
)

type tracingImpl struct {
	channelOwner
	includeSources    bool
	isTracing         bool
	isLive            bool
	stacksId          string
	tracesDir         string
	harRecorders      map[string]harRecordingMetadata
	additionalSources map[string]struct{}
}

func (t *tracingImpl) Start(options ...TracingStartOptions) error {
	chunkOption := TracingStartChunkOptions{}
	if len(options) == 1 {
		if options[0].Sources != nil {
			t.includeSources = *options[0].Sources
		}
		t.isLive = options[0].Live != nil && *options[0].Live
		chunkOption.Name = options[0].Name
		chunkOption.Title = options[0].Title
	}
	innerStart := func() (any, error) {
		if _, err := t.channel.Send("tracingStart", options); err != nil {
			return "", err
		}
		return t.channel.Send("tracingStartChunk", chunkOption)
	}
	name, err := innerStart()
	if err != nil {
		return err
	}
	return t.startCollectingStacks(name.(string))
}

func (t *tracingImpl) StartChunk(options ...TracingStartChunkOptions) error {
	name, err := t.channel.Send("tracingStartChunk", options)
	if err != nil {
		return err
	}
	return t.startCollectingStacks(name.(string))
}

func (t *tracingImpl) StopChunk(path ...string) error {
	filePath := ""
	if len(path) == 1 {
		filePath = path[0]
	}
	return t.doStopChunk(filePath)
}

func (t *tracingImpl) Stop(path ...string) error {
	filePath := ""
	if len(path) == 1 {
		filePath = path[0]
	}
	if err := t.doStopChunk(filePath); err != nil {
		return err
	}
	_, err := t.channel.Send("tracingStop")
	return err
}

// resetStackCounter clears the in-tracing flag and decrements the connection
// tracing counter, mirroring upstream _resetStackCounter. Called on context
// close so an un-stopped trace doesn't keep the connection collecting stacks.
func (t *tracingImpl) resetStackCounter() {
	if t.isTracing {
		t.isTracing = false
		t.connection.setInTracing(false)
	}
}

func (t *tracingImpl) doStopChunk(filePath string) (err error) {
	if t.isTracing {
		t.isTracing = false
		t.connection.setInTracing(false)
	}

	additionalSources := make([]string, 0, len(t.additionalSources))
	for source := range t.additionalSources {
		additionalSources = append(additionalSources, source)
	}
	t.additionalSources = make(map[string]struct{})

	if filePath == "" {
		// Not interested in artifacts.
		_, err = t.channel.Send("tracingStopChunk", map[string]any{
			"mode": "discard",
		})
		if t.stacksId != "" {
			return t.connection.LocalUtils().TraceDiscarded(t.stacksId)
		}
		return err
	}

	isLocal := !t.connection.isRemote
	if isLocal {
		result, err := t.channel.SendReturnAsDict("tracingStopChunk", map[string]any{
			"mode": "entries",
		})
		if err != nil {
			return err
		}
		entries, ok := result["entries"]
		if !ok {
			return fmt.Errorf("could not convert result to map: %v", result)
		}
		_, err = t.connection.LocalUtils().Zip(localUtilsZipOptions{
			ZipFile:           filePath,
			Entries:           entries.([]any),
			StacksId:          t.stacksId,
			Mode:              "write",
			IncludeSources:    t.includeSources,
			AdditionalSources: additionalSources,
		})
		return err
	}

	result, err := t.channel.SendReturnAsDict("tracingStopChunk", map[string]any{
		"mode": "archive",
	})
	if err != nil {
		return err
	}
	artifactChannel, ok := result["artifact"]
	if !ok {
		return fmt.Errorf("could not convert result to map: %v", result)
	}
	// Save trace to the final local file.
	artifact := fromNullableChannel(artifactChannel).(*artifactImpl)
	// The artifact may be missing if the browser closed while stopping tracing.
	if artifact == nil {
		if t.stacksId != "" {
			return t.connection.LocalUtils().TraceDiscarded(t.stacksId)
		}
		return
	}
	if err := artifact.SaveAs(filePath); err != nil {
		return err
	}
	if err := artifact.Delete(); err != nil {
		return err
	}
	_, err = t.connection.LocalUtils().Zip(localUtilsZipOptions{
		ZipFile:           filePath,
		Entries:           []any{},
		StacksId:          t.stacksId,
		Mode:              "append",
		IncludeSources:    t.includeSources,
		AdditionalSources: additionalSources,
	})
	return err
}

func (t *tracingImpl) startCollectingStacks(name string) (err error) {
	if !t.isTracing {
		t.isTracing = true
		t.connection.setInTracing(true)
	}
	t.stacksId, err = t.connection.LocalUtils().TracingStarted(name, t.isLive, t.tracesDir)
	return
}

func (t *tracingImpl) Group(name string, options ...TracingGroupOptions) error {
	var option TracingGroupOptions
	if len(options) == 1 {
		option = options[0]
		if option.Location != nil && option.Location.File != "" {
			t.additionalSources[option.Location.File] = struct{}{}
		}
	}
	_, err := t.channel.Send("tracingGroup", option, map[string]any{"name": name})
	return err
}

func (t *tracingImpl) GroupEnd() error {
	_, err := t.channel.Send("tracingGroupEnd")
	return err
}

func (t *tracingImpl) StartHar(path string, options ...TracingStartHarOptions) error {
	if len(t.harRecorders) > 0 {
		return fmt.Errorf("HAR recording has already been started")
	}
	isZip := strings.HasSuffix(strings.ToLower(path), ".zip")
	// Default content matches upstream: attach for .zip output, embed otherwise.
	defaultContent := HarContentPolicyEmbed
	if isZip {
		defaultContent = HarContentPolicyAttach
	}
	harOptions := recordHarInputOptions{
		Path:    path,
		Content: defaultContent,
		Mode:    HarModeFull,
	}
	if len(options) == 1 {
		if options[0].ResourcesDir != nil && isZip {
			return fmt.Errorf("resourcesDir option is not compatible with a .zip har file")
		}
		if options[0].Content != nil {
			harOptions.Content = options[0].Content
		}
		if options[0].Mode != nil {
			harOptions.Mode = options[0].Mode
		}
		harOptions.URL = options[0].URLFilter
		harOptions.ResourcesDir = options[0].ResourcesDir
	}
	harId, err := t.channel.Send("harStart", map[string]any{
		"options": prepareRecordHarOptions(harOptions),
	})
	if err != nil {
		return err
	}
	t.harRecorders[harId.(string)] = harRecordingMetadata{
		Path:         path,
		Content:      harOptions.Content,
		ResourcesDir: harOptions.ResourcesDir,
	}
	return nil
}

func (t *tracingImpl) StopHar() error {
	if len(t.harRecorders) == 0 {
		return fmt.Errorf("HAR recording has not been started")
	}
	return t.exportAllHars()
}

// exportAllHars flushes every active HAR recording to disk. It is invoked by
// StopHar and by APIRequestContext.Dispose so HARs started via the request
// context are written even without an explicit StopHar call.
func (t *tracingImpl) exportAllHars() error {
	for harId, harMetaData := range t.harRecorders {
		delete(t.harRecorders, harId)
		overrides := map[string]any{}
		if harId != "" {
			overrides["harId"] = harId
		}
		needCompressed := strings.HasSuffix(strings.ToLower(harMetaData.Path), ".zip")
		if !t.connection.isRemote {
			overrides["mode"] = "entries"
			response, err := t.channel.SendReturnAsDict("harExport", overrides)
			if err != nil {
				return err
			}
			if !needCompressed {
				continue
			}
			entries, ok := response["entries"].([]any)
			if !ok {
				return fmt.Errorf("could not convert HAR entries: %v", response)
			}
			if _, err = t.connection.LocalUtils().Zip(localUtilsZipOptions{
				ZipFile: harMetaData.Path,
				Entries: entries,
				Mode:    "write",
			}); err != nil {
				return err
			}
			continue
		}
		overrides["mode"] = "archive"
		response, err := t.channel.SendReturnAsDict("harExport", overrides)
		if err != nil {
			return err
		}
		artifact := fromChannel(response["artifact"]).(*artifactImpl)
		if needCompressed {
			if err := artifact.SaveAs(harMetaData.Path); err != nil {
				return err
			}
		} else {
			tmpPath := harMetaData.Path + ".tmp"
			if err := artifact.SaveAs(tmpPath); err != nil {
				return err
			}
			if err := t.connection.localUtils.HarUnzip(tmpPath, harMetaData.Path, harMetaData.ResourcesDir); err != nil {
				return err
			}
		}
		if err := artifact.Delete(); err != nil {
			return err
		}
	}
	return nil
}

func newTracing(parent *channelOwner, objectType string, guid string, initializer map[string]any) *tracingImpl {
	bt := &tracingImpl{
		harRecorders:      make(map[string]harRecordingMetadata),
		additionalSources: make(map[string]struct{}),
	}
	bt.createChannelOwner(bt, parent, objectType, guid, initializer)
	bt.markAsInternalType()
	return bt
}
