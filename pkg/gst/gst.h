#ifndef GST_H
#define GST_H

#include <glib.h>
#include <gst/gst.h>
#include <gst/app/gstappsink.h>
#include <gst/app/gstappsrc.h>
#include <gst/video/video.h>

typedef struct GstPipelineCtx {
  GstElement *pipeline;
  GstElement *appsink;
  GstElement *appsrc;
  int id;
} GstPipelineCtx;

GstPipelineCtx* gstreamer_pipeline_create(char *pipeline, int id, GError **error);
void gstreamer_pipeline_play(GstPipelineCtx *ctx);
void gstreamer_pipeline_pause(GstPipelineCtx *ctx);
void gstreamer_pipeline_destory(GstPipelineCtx *ctx);
void gstreamer_pipeline_attach_appsink(GstPipelineCtx *ctx, char *sink_name);
void gstreamer_pipeline_emit_video_keyframe(GstPipelineCtx *ctx);

// Callback declarations for Go
extern void goHandlePipelineBuffer(int id, void *buffer, int buffer_len, int duration, int is_keyframe);

#endif
