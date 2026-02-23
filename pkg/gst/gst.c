#include "gst.h"

static GstFlowReturn gstreamer_send_new_sample_handler(GstElement *object, gpointer user_data) {
  GstSample *sample = NULL;
  GstBuffer *buffer = NULL;
  GstPipelineCtx *ctx = (GstPipelineCtx*)user_data;

  g_signal_emit_by_name(object, "pull-sample", &sample);
  if (sample) {
    buffer = gst_sample_get_buffer(sample);
    if (buffer) {
      GstMapInfo map;
      gst_buffer_map(buffer, &map, GST_MAP_READ);

      int is_keyframe = !GST_BUFFER_FLAG_IS_SET(buffer, GST_BUFFER_FLAG_DELTA_UNIT);
      goHandlePipelineBuffer(ctx->id, map.data, map.size, GST_BUFFER_DURATION(buffer), is_keyframe);

      gst_buffer_unmap(buffer, &map);
    }
    gst_sample_unref(sample);
  }

  return GST_FLOW_OK;
}

GstPipelineCtx* gstreamer_pipeline_create(char *pipeline, int id, GError **error) {
  GstElement *gst_pipeline = gst_parse_launch(pipeline, error);
  if (*error != NULL) {
    return NULL;
  }

  GstPipelineCtx *ctx = calloc(1, sizeof(GstPipelineCtx));
  ctx->pipeline = gst_pipeline;
  ctx->id = id;
  return ctx;
}

void gstreamer_pipeline_attach_appsink(GstPipelineCtx *ctx, char *sink_name) {
  GstElement *appsink = gst_bin_get_by_name(GST_BIN(ctx->pipeline), sink_name);
  ctx->appsink = appsink;

  g_object_set(appsink, "emit-signals", TRUE, "sync", FALSE, NULL);
  g_signal_connect(appsink, "new-sample", G_CALLBACK(gstreamer_send_new_sample_handler), ctx);
}

void gstreamer_pipeline_play(GstPipelineCtx *ctx) {
  gst_element_set_state(ctx->pipeline, GST_STATE_PLAYING);
}

void gstreamer_pipeline_pause(GstPipelineCtx *ctx) {
  gst_element_set_state(ctx->pipeline, GST_STATE_PAUSED);
}

void gstreamer_pipeline_destory(GstPipelineCtx *ctx) {
  gst_element_set_state(ctx->pipeline, GST_STATE_NULL);
  gst_object_unref(ctx->pipeline);
  free(ctx);
}

void gstreamer_pipeline_emit_video_keyframe(GstPipelineCtx *ctx) {
  GstPad *sinkpad = gst_element_get_static_pad(ctx->pipeline, "encoder");
  if (sinkpad) {
    gst_pad_push_event(sinkpad, gst_video_event_new_upstream_force_key_unit(GST_CLOCK_TIME_NONE, TRUE, 0));
    gst_object_unref(sinkpad);
  }
}
