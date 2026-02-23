package playwright

import (
	"errors"
	"fmt"
)

type artifactImpl struct {
	channelOwner
}

func (a *artifactImpl) AbsolutePath() string {
	return a.initializer["absolutePath"].(string)
}

func (a *artifactImpl) PathAfterFinished() (string, error) {
	if a.connection.isRemote {
		return "", errors.New("Path is not available when connecting remotely. Use SaveAs() to save a local copy")
	}
	path, err := a.channel.Send("pathAfterFinished")
	return path.(string), err
}

func (a *artifactImpl) SaveAs(path string) error {
	if !a.connection.isRemote {
		_, err := a.channel.Send("saveAs", map[string]any{
			"path": path,
		})
		return err
	}
	streamChannel, err := a.channel.Send("saveAsStream")
	if err != nil {
		return err
	}
	stream := fromChannel(streamChannel).(*streamImpl)
	return stream.SaveAs(path)
}

func (a *artifactImpl) Failure() error {
	reason, err := a.channel.Send("failure")
	if reason == nil {
		return err
	}
	return fmt.Errorf("%w: %v", ErrPlaywright, reason)
}

func (a *artifactImpl) Delete() error {
	_, err := a.channel.Send("delete")
	return err
}

func (a *artifactImpl) Cancel() error {
	_, err := a.channel.Send("cancel")
	return err
}

func (a *artifactImpl) ReadIntoBuffer() ([]byte, error) {
	streamChannel, err := a.channel.Send("stream")
	if err != nil {
		return nil, err
	}
	stream := fromChannel(streamChannel)
	return stream.(*streamImpl).ReadAll()
}

func newArtifact(parent *channelOwner, objectType string, guid string, initializer map[string]any) *artifactImpl {
	artifact := &artifactImpl{}
	artifact.createChannelOwner(artifact, parent, objectType, guid, initializer)
	return artifact
}
