//go:build integration

package test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/livekit/egress/pkg/types"
	"github.com/livekit/protocol/livekit"
	"github.com/livekit/protocol/utils"
)

func testRoomCompositeFile(t *testing.T, conf *TestConfig) {
	publishSamplesToRoom(t, conf.room, types.MimeTypeOpus, types.MimeTypeH264, conf.Muting)

	for _, test := range []*testCase{
		{
			name:     "h264-high-mp4",
			fileType: livekit.EncodedFileType_MP4,
			options: &livekit.EncodingOptions{
				AudioCodec: livekit.AudioCodec_AAC,
				VideoCodec: livekit.VideoCodec_H264_HIGH,
			},
			filename: "r_{room_name}_high_{time}.mp4",
		},
		{
			name:     "h264-high-mp4-limit",
			fileType: livekit.EncodedFileType_MP4,
			options: &livekit.EncodingOptions{
				AudioCodec:   livekit.AudioCodec_AAC,
				Width:        1280,
				Height:       720,
				VideoBitrate: 4500,
			},
			filename:       "r_limit_{time}.mp4",
			sessionTimeout: time.Second * 20,
		},
		{
			name:      "opus-ogg",
			fileType:  livekit.EncodedFileType_OGG,
			audioOnly: true,
			options: &livekit.EncodingOptions{
				AudioCodec: livekit.AudioCodec_OPUS,
			},
			filename: "r_{room_name}_opus_{time}",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			awaitIdle(t, conf.svc)

			roomRequest := &livekit.RoomCompositeEgressRequest{
				RoomName:  conf.room.Name(),
				Layout:    "speaker-dark",
				AudioOnly: test.audioOnly,
				Output: &livekit.RoomCompositeEgressRequest_File{
					File: &livekit.EncodedFileOutput{
						FileType: test.fileType,
						Filepath: getFilePath(conf.ServiceConfig, test.filename),
					},
				},
			}

			if test.options != nil {
				roomRequest.Options = &livekit.RoomCompositeEgressRequest_Advanced{
					Advanced: test.options,
				}
			}

			req := &livekit.StartEgressRequest{
				EgressId: utils.NewGuid(utils.EgressPrefix),
				Request: &livekit.StartEgressRequest_RoomComposite{
					RoomComposite: roomRequest,
				},
			}

			runFileTest(t, conf, req, test)
		})
	}
}

func testRoomCompositeStream(t *testing.T, conf *TestConfig) {
	publishSamplesToRoom(t, conf.room, types.MimeTypeOpus, types.MimeTypeVP8, conf.Muting)

	t.Run("rtmp-failure", func(t *testing.T) {
		awaitIdle(t, conf.svc)

		req := &livekit.StartEgressRequest{
			EgressId: utils.NewGuid(utils.EgressPrefix),
			Request: &livekit.StartEgressRequest_RoomComposite{
				RoomComposite: &livekit.RoomCompositeEgressRequest{
					RoomName: conf.RoomName,
					Layout:   "speaker-light",
					Output: &livekit.RoomCompositeEgressRequest_Stream{
						Stream: &livekit.StreamOutput{
							Protocol: livekit.StreamProtocol_RTMP,
							Urls:     []string{badStreamUrl},
						},
					},
				},
			},
		}

		info, err := conf.rpcClient.SendRequest(context.Background(), req)
		require.NoError(t, err)
		require.Empty(t, info.Error)
		require.NotEmpty(t, info.EgressId)
		require.Equal(t, conf.RoomName, info.RoomName)
		require.Equal(t, livekit.EgressStatus_EGRESS_STARTING, info.Status)

		// wait
		time.Sleep(time.Second * 5)

		info = getUpdate(t, conf.updates, info.EgressId)
		if info.Status == livekit.EgressStatus_EGRESS_ACTIVE {
			checkUpdate(t, conf.updates, info.EgressId, livekit.EgressStatus_EGRESS_FAILED)
		} else {
			require.Equal(t, info.Status, livekit.EgressStatus_EGRESS_FAILED)
		}
	})

	for _, test := range []*testCase{
		{
			name: "room-rtmp",
		},
		{
			name:           "room-rtmp-limit",
			sessionTimeout: time.Second * 20,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			awaitIdle(t, conf.svc)

			req := &livekit.StartEgressRequest{
				EgressId: utils.NewGuid(utils.EgressPrefix),
				Request: &livekit.StartEgressRequest_RoomComposite{
					RoomComposite: &livekit.RoomCompositeEgressRequest{
						RoomName: conf.room.Name(),
						Layout:   "grid-light",
						Output: &livekit.RoomCompositeEgressRequest_Stream{
							Stream: &livekit.StreamOutput{
								Protocol: livekit.StreamProtocol_RTMP,
								Urls:     []string{streamUrl1},
							},
						},
					},
				},
			}

			runStreamTest(t, conf, req, test.sessionTimeout)
		})
	}
}

func testRoomCompositeSegments(t *testing.T, conf *TestConfig) {
	publishSamplesToRoom(t, conf.room, types.MimeTypeOpus, types.MimeTypeVP8, conf.Muting)

	for _, test := range []*testCase{
		{
			name: "rs-baseline",
			options: &livekit.EncodingOptions{
				AudioCodec:   livekit.AudioCodec_AAC,
				VideoCodec:   livekit.VideoCodec_H264_BASELINE,
				Width:        1920,
				Height:       1080,
				VideoBitrate: 4500,
			},
			filename: "rs_{room_name}_{time}",
			playlist: "rs_{room_name}_{time}.m3u8",
		},
		{
			name: "rs-limit",
			options: &livekit.EncodingOptions{
				AudioCodec:   livekit.AudioCodec_AAC,
				VideoCodec:   livekit.VideoCodec_H264_BASELINE,
				Width:        1920,
				Height:       1080,
				VideoBitrate: 4500,
			},
			filename:       "rs_limit_{time}",
			playlist:       "rs_limit_{time}.m3u8",
			sessionTimeout: time.Second * 20,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			awaitIdle(t, conf.svc)

			roomRequest := &livekit.RoomCompositeEgressRequest{
				RoomName:  conf.RoomName,
				Layout:    "grid-dark",
				AudioOnly: test.audioOnly,
				Output: &livekit.RoomCompositeEgressRequest_Segments{
					Segments: &livekit.SegmentedFileOutput{
						FilenamePrefix: getFilePath(conf.ServiceConfig, test.filename),
						PlaylistName:   test.playlist,
					},
				},
			}

			if test.options != nil {
				roomRequest.Options = &livekit.RoomCompositeEgressRequest_Advanced{
					Advanced: test.options,
				}
			}

			req := &livekit.StartEgressRequest{
				EgressId: utils.NewGuid(utils.EgressPrefix),
				Request: &livekit.StartEgressRequest_RoomComposite{
					RoomComposite: roomRequest,
				},
			}

			runSegmentsTest(t, conf, req, test.sessionTimeout)
		})
		return
	}
}
