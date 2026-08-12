package playwright

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAllActionOptionsSerializeScrollMode(t *testing.T) {
	options := []any{
		ElementHandleCheckOptions{Scroll: ScrollModeAuto},
		ElementHandleClickOptions{Scroll: ScrollModeAuto},
		ElementHandleDblclickOptions{Scroll: ScrollModeAuto},
		ElementHandleHoverOptions{Scroll: ScrollModeAuto},
		ElementHandleSetCheckedOptions{Scroll: ScrollModeAuto},
		ElementHandleTapOptions{Scroll: ScrollModeAuto},
		ElementHandleUncheckOptions{Scroll: ScrollModeAuto},
		FrameCheckOptions{Scroll: ScrollModeAuto},
		FrameClickOptions{Scroll: ScrollModeAuto},
		FrameDblclickOptions{Scroll: ScrollModeAuto},
		FrameDragAndDropOptions{Scroll: ScrollModeAuto},
		FrameHoverOptions{Scroll: ScrollModeAuto},
		FrameSetCheckedOptions{Scroll: ScrollModeAuto},
		FrameTapOptions{Scroll: ScrollModeAuto},
		FrameUncheckOptions{Scroll: ScrollModeAuto},
		LocatorCheckOptions{Scroll: ScrollModeAuto},
		LocatorClickOptions{Scroll: ScrollModeAuto},
		LocatorDblclickOptions{Scroll: ScrollModeAuto},
		LocatorDragToOptions{Scroll: ScrollModeAuto},
		LocatorHoverOptions{Scroll: ScrollModeAuto},
		LocatorSetCheckedOptions{Scroll: ScrollModeAuto},
		LocatorTapOptions{Scroll: ScrollModeAuto},
		LocatorUncheckOptions{Scroll: ScrollModeAuto},
		PageCheckOptions{Scroll: ScrollModeAuto},
		PageClickOptions{Scroll: ScrollModeAuto},
		PageDblclickOptions{Scroll: ScrollModeAuto},
		PageDragAndDropOptions{Scroll: ScrollModeAuto},
		PageHoverOptions{Scroll: ScrollModeAuto},
		PageSetCheckedOptions{Scroll: ScrollModeAuto},
		PageTapOptions{Scroll: ScrollModeAuto},
		PageUncheckOptions{Scroll: ScrollModeAuto},
	}

	for _, option := range options {
		encoded, err := json.Marshal(option)
		require.NoError(t, err)
		var payload map[string]any
		require.NoError(t, json.Unmarshal(encoded, &payload))
		require.Equal(t, "auto", payload["scroll"])
	}
}
