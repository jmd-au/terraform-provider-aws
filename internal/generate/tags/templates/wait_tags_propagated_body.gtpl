// {{ .WaitTagsPropagatedFunc }} waits for {{ .ServicePackage }} service tags to be propagated.
// The identifier is typically the Amazon Resource Name (ARN), although
// it may also be a different identifier depending on the service.
func {{ .WaitTagsPropagatedFunc }}(ctx context.Context, conn {{ .ClientType }}, id string, tags tftags.KeyValueTags, optFns ...func(*{{ .AWSService }}.Options)) error {
	tflog.Debug(ctx, "Waiting for tag propagation", map[string]any{
		names.AttrTags: tags,
	})

	checkFunc := func(ctx context.Context) (bool, error) {
		output, err := {{ .ListTagsFunc }}(ctx, conn, id, optFns...)

		if tfresource.NotFound(err) {
			return false, nil
		}

		if err != nil {
			return false, smarterr.NewError(err)
		}

		if inContext, ok := tftags.FromContext(ctx); ok {
			tags = tags.IgnoreConfig(inContext.IgnoreConfig)
			output = output.IgnoreConfig(inContext.IgnoreConfig)
		}

        return output.{{ .WaitFuncComparator }}(tags), nil
	}
	opts := tfresource.WaitOpts{
		{{- if ne .WaitContinuousOccurence 0 }}
		ContinuousTargetOccurence: {{ .WaitContinuousOccurence }},
		{{- end }}
		{{- if ne .WaitDelay "" }}
		Delay: {{ .WaitDelay }},
		{{- end }}
		{{- if ne .WaitMinTimeout "" }}
		MinTimeout: {{ .WaitMinTimeout }},
		{{- end }}
		{{- if ne .WaitPollInterval "" }}
		PollInterval: {{ .WaitPollInterval }},
		{{- end }}
	}

	return tfresource.WaitUntil(ctx, {{ .WaitTimeout }}, checkFunc, opts)
}
