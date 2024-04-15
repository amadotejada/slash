package v1

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	apiv1pb "github.com/yourselfhosted/slash/proto/gen/api/v1"
	storepb "github.com/yourselfhosted/slash/proto/gen/store"
	"github.com/yourselfhosted/slash/store"
)

func (s *APIV1Service) GetWorkspaceProfile(ctx context.Context, _ *apiv1pb.GetWorkspaceProfileRequest) (*apiv1pb.GetWorkspaceProfileResponse, error) {
	profile := &apiv1pb.WorkspaceProfile{
		Mode:    s.Profile.Mode,
		Version: s.Profile.Version,
		Plan:    apiv1pb.PlanType_FREE,
	}

	// Load subscription plan from license service.
	subscription, err := s.LicenseService.GetSubscription(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get subscription: %v", err)
	}
	profile.Plan = subscription.Plan

	workspaceSetting, err := s.GetWorkspaceSetting(ctx, &apiv1pb.GetWorkspaceSettingRequest{})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get workspace setting: %v", err)
	}
	if workspaceSetting != nil {
		setting := workspaceSetting.GetSetting()
		profile.EnableSignup = setting.GetEnableSignup()
		profile.CustomStyle = setting.GetCustomStyle()
		profile.CustomScript = setting.GetCustomScript()
		profile.FaviconProvider = setting.GetFaviconProvider()
	}
	return &apiv1pb.GetWorkspaceProfileResponse{
		Profile: profile,
	}, nil
}

func (s *APIV1Service) GetWorkspaceSetting(ctx context.Context, _ *apiv1pb.GetWorkspaceSettingRequest) (*apiv1pb.GetWorkspaceSettingResponse, error) {
	isAdmin := false
	userID, ok := ctx.Value(userIDContextKey).(int32)
	if ok {
		user, err := s.Store.GetUser(ctx, &store.FindUser{ID: &userID})
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to get user: %v", err)
		}
		if user.Role == store.RoleAdmin {
			isAdmin = true
		}
	}
	workspaceSettings, err := s.Store.ListWorkspaceSettings(ctx, &store.FindWorkspaceSetting{})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list workspace settings: %v", err)
	}
	workspaceSetting := &apiv1pb.WorkspaceSetting{
		EnableSignup: true,
	}
	for _, v := range workspaceSettings {
		if v.Key == storepb.WorkspaceSettingKey_WORKSAPCE_SETTING_ENABLE_SIGNUP {
			workspaceSetting.EnableSignup = v.GetEnableSignup()
		} else if v.Key == storepb.WorkspaceSettingKey_WORKSPACE_SETTING_INSTANCE_URL {
			workspaceSetting.InstanceUrl = v.GetInstanceUrl()
		} else if v.Key == storepb.WorkspaceSettingKey_WORKSPACE_SETTING_CUSTOM_STYLE {
			workspaceSetting.CustomStyle = v.GetCustomStyle()
		} else if v.Key == storepb.WorkspaceSettingKey_WORKSPACE_SETTING_CUSTOM_SCRIPT {
			workspaceSetting.CustomScript = v.GetCustomScript()
		} else if v.Key == storepb.WorkspaceSettingKey_WORKSPACE_SETTING_DEFAULT_VISIBILITY {
			workspaceSetting.DefaultVisibility = apiv1pb.Visibility(v.GetDefaultVisibility())
		} else if v.Key == storepb.WorkspaceSettingKey_WORKSPACE_SETTING_FAVICON_PROVIDER {
			workspaceSetting.FaviconProvider = v.GetFaviconProvider()
		} else if isAdmin {
			// For some settings, only admin can get the value.
			if v.Key == storepb.WorkspaceSettingKey_WORKSPACE_SETTING_LICENSE_KEY {
				workspaceSetting.LicenseKey = v.GetLicenseKey()
			}
		}
	}
	return &apiv1pb.GetWorkspaceSettingResponse{
		Setting: workspaceSetting,
	}, nil
}

func (s *APIV1Service) UpdateWorkspaceSetting(ctx context.Context, request *apiv1pb.UpdateWorkspaceSettingRequest) (*apiv1pb.UpdateWorkspaceSettingResponse, error) {
	if request.UpdateMask == nil || len(request.UpdateMask.Paths) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "update mask is empty")
	}

	for _, path := range request.UpdateMask.Paths {
		if path == "license_key" {
			if _, err := s.Store.UpsertWorkspaceSetting(ctx, &storepb.WorkspaceSetting{
				Key: storepb.WorkspaceSettingKey_WORKSPACE_SETTING_LICENSE_KEY,
				Value: &storepb.WorkspaceSetting_LicenseKey{
					LicenseKey: request.Setting.LicenseKey,
				},
			}); err != nil {
				return nil, status.Errorf(codes.Internal, "failed to update workspace setting: %v", err)
			}
		} else if path == "enable_signup" {
			if _, err := s.Store.UpsertWorkspaceSetting(ctx, &storepb.WorkspaceSetting{
				Key: storepb.WorkspaceSettingKey_WORKSAPCE_SETTING_ENABLE_SIGNUP,
				Value: &storepb.WorkspaceSetting_EnableSignup{
					EnableSignup: request.Setting.EnableSignup,
				},
			}); err != nil {
				return nil, status.Errorf(codes.Internal, "failed to update workspace setting: %v", err)
			}
		} else if path == "instance_url" {
			if _, err := s.Store.UpsertWorkspaceSetting(ctx, &storepb.WorkspaceSetting{
				Key: storepb.WorkspaceSettingKey_WORKSPACE_SETTING_INSTANCE_URL,
				Value: &storepb.WorkspaceSetting_InstanceUrl{
					InstanceUrl: request.Setting.InstanceUrl,
				},
			}); err != nil {
				return nil, status.Errorf(codes.Internal, "failed to update workspace setting: %v", err)
			}
		} else if path == "custom_style" {
			if _, err := s.Store.UpsertWorkspaceSetting(ctx, &storepb.WorkspaceSetting{
				Key: storepb.WorkspaceSettingKey_WORKSPACE_SETTING_CUSTOM_STYLE,
				Value: &storepb.WorkspaceSetting_CustomStyle{
					CustomStyle: request.Setting.CustomStyle,
				},
			}); err != nil {
				return nil, status.Errorf(codes.Internal, "failed to update workspace setting: %v", err)
			}
		} else if path == "custom_script" {
			if _, err := s.Store.UpsertWorkspaceSetting(ctx, &storepb.WorkspaceSetting{
				Key: storepb.WorkspaceSettingKey_WORKSPACE_SETTING_CUSTOM_SCRIPT,
				Value: &storepb.WorkspaceSetting_CustomScript{
					CustomScript: request.Setting.CustomScript,
				},
			}); err != nil {
				return nil, status.Errorf(codes.Internal, "failed to update workspace setting: %v", err)
			}
		} else if path == "default_visibility" {
			if _, err := s.Store.UpsertWorkspaceSetting(ctx, &storepb.WorkspaceSetting{
				Key: storepb.WorkspaceSettingKey_WORKSPACE_SETTING_DEFAULT_VISIBILITY,
				Value: &storepb.WorkspaceSetting_DefaultVisibility{
					DefaultVisibility: storepb.Visibility(request.Setting.DefaultVisibility),
				},
			}); err != nil {
				return nil, status.Errorf(codes.Internal, "failed to update workspace setting: %v", err)
			}
		} else if path == "favicon_provider" {
			if _, err := s.Store.UpsertWorkspaceSetting(ctx, &storepb.WorkspaceSetting{
				Key: storepb.WorkspaceSettingKey_WORKSPACE_SETTING_FAVICON_PROVIDER,
				Value: &storepb.WorkspaceSetting_FaviconProvider{
					FaviconProvider: request.Setting.FaviconProvider,
				},
			}); err != nil {
				return nil, status.Errorf(codes.Internal, "failed to update workspace setting: %v", err)
			}
		} else {
			return nil, status.Errorf(codes.InvalidArgument, "invalid path: %s", path)
		}
	}

	getWorkspaceSettingResponse, err := s.GetWorkspaceSetting(ctx, &apiv1pb.GetWorkspaceSettingRequest{})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get workspace setting: %v", err)
	}
	return &apiv1pb.UpdateWorkspaceSettingResponse{
		Setting: getWorkspaceSettingResponse.Setting,
	}, nil
}
