// @generated by protoc-gen-es v2.2.3 with parameter "target=ts"
// @generated from file user/v1/auth.proto (package user.v1, syntax proto3)
/* eslint-disable */

import type { GenFile, GenMessage, GenService } from "@bufbuild/protobuf/codegenv1";
import { fileDesc, messageDesc, serviceDesc } from "@bufbuild/protobuf/codegenv1";
import type { Message } from "@bufbuild/protobuf";

/**
 * Describes the file user/v1/auth.proto.
 */
export const file_user_v1_auth: GenFile = /*@__PURE__*/
  fileDesc("ChJ1c2VyL3YxL2F1dGgucHJvdG8SB3VzZXIudjEiMgoMTG9naW5SZXF1ZXN0EhAKCHVzZXJuYW1lGAEgASgJEhAKCHBhc3N3b3JkGAIgASgJIh4KDUxvZ2luUmVzcG9uc2USDQoFdG9rZW4YASABKAkiTQoNU2lnblVwUmVxdWVzdBIQCgh1c2VybmFtZRgBIAEoCRIQCghwYXNzd29yZBgCIAEoCRIYChBjb25maXJtX3Bhc3N3b3JkGAMgASgJIhAKDlNpZ25VcFJlc3BvbnNlIg8KDUxvZ291dFJlcXVlc3QiEAoOTG9nb3V0UmVzcG9uc2UiKAoUR2V0UGFzc2tleUlEc1JlcXVlc3QSEAoIdXNlcm5hbWUYASABKAkiLAoVR2V0UGFzc2tleUlEc1Jlc3BvbnNlEhMKC3Bhc3NrZXlfaWRzGAEgAygJIhoKGEJlZ2luUGFzc2tleUxvZ2luUmVxdWVzdCIbChlCZWdpblBhc3NrZXlMb2dpblJlc3BvbnNlIhsKGUZpbmlzaFBhc3NrZXlMb2dpblJlcXVlc3QiHAoaRmluaXNoUGFzc2tleUxvZ2luUmVzcG9uc2Uy0gMKC0F1dGhTZXJ2aWNlEjgKBUxvZ2luEhUudXNlci52MS5Mb2dpblJlcXVlc3QaFi51c2VyLnYxLkxvZ2luUmVzcG9uc2UiABI7CgZTaWduVXASFi51c2VyLnYxLlNpZ25VcFJlcXVlc3QaFy51c2VyLnYxLlNpZ25VcFJlc3BvbnNlIgASOwoGTG9nb3V0EhYudXNlci52MS5Mb2dvdXRSZXF1ZXN0GhcudXNlci52MS5Mb2dvdXRSZXNwb25zZSIAElAKDUdldFBhc3NrZXlJRHMSHS51c2VyLnYxLkdldFBhc3NrZXlJRHNSZXF1ZXN0Gh4udXNlci52MS5HZXRQYXNza2V5SURzUmVzcG9uc2UiABJcChFCZWdpblBhc3NrZXlMb2dpbhIhLnVzZXIudjEuQmVnaW5QYXNza2V5TG9naW5SZXF1ZXN0GiIudXNlci52MS5CZWdpblBhc3NrZXlMb2dpblJlc3BvbnNlIgASXwoSRmluaXNoUGFzc2tleUxvZ2luEiIudXNlci52MS5GaW5pc2hQYXNza2V5TG9naW5SZXF1ZXN0GiMudXNlci52MS5GaW5pc2hQYXNza2V5TG9naW5SZXNwb25zZSIAQp0BCgtjb20udXNlci52MUIJQXV0aFByb3RvUAFaRmdpdGh1Yi5jb20vc3BvdGRlbW80L3RyZXZzdGFjay9zZXJ2ZXIvaW50ZXJuYWwvc2VydmljZXMvdXNlci92MTt1c2VydjGiAgNVWFiqAgdVc2VyLlYxygIHVXNlclxWMeICE1VzZXJcVjFcR1BCTWV0YWRhdGHqAghVc2VyOjpWMWIGcHJvdG8z");

/**
 * @generated from message user.v1.LoginRequest
 */
export type LoginRequest = Message<"user.v1.LoginRequest"> & {
  /**
   * @generated from field: string username = 1;
   */
  username: string;

  /**
   * @generated from field: string password = 2;
   */
  password: string;
};

/**
 * Describes the message user.v1.LoginRequest.
 * Use `create(LoginRequestSchema)` to create a new message.
 */
export const LoginRequestSchema: GenMessage<LoginRequest> = /*@__PURE__*/
  messageDesc(file_user_v1_auth, 0);

/**
 * @generated from message user.v1.LoginResponse
 */
export type LoginResponse = Message<"user.v1.LoginResponse"> & {
  /**
   * @generated from field: string token = 1;
   */
  token: string;
};

/**
 * Describes the message user.v1.LoginResponse.
 * Use `create(LoginResponseSchema)` to create a new message.
 */
export const LoginResponseSchema: GenMessage<LoginResponse> = /*@__PURE__*/
  messageDesc(file_user_v1_auth, 1);

/**
 * @generated from message user.v1.SignUpRequest
 */
export type SignUpRequest = Message<"user.v1.SignUpRequest"> & {
  /**
   * @generated from field: string username = 1;
   */
  username: string;

  /**
   * @generated from field: string password = 2;
   */
  password: string;

  /**
   * @generated from field: string confirm_password = 3;
   */
  confirmPassword: string;
};

/**
 * Describes the message user.v1.SignUpRequest.
 * Use `create(SignUpRequestSchema)` to create a new message.
 */
export const SignUpRequestSchema: GenMessage<SignUpRequest> = /*@__PURE__*/
  messageDesc(file_user_v1_auth, 2);

/**
 * @generated from message user.v1.SignUpResponse
 */
export type SignUpResponse = Message<"user.v1.SignUpResponse"> & {
};

/**
 * Describes the message user.v1.SignUpResponse.
 * Use `create(SignUpResponseSchema)` to create a new message.
 */
export const SignUpResponseSchema: GenMessage<SignUpResponse> = /*@__PURE__*/
  messageDesc(file_user_v1_auth, 3);

/**
 * @generated from message user.v1.LogoutRequest
 */
export type LogoutRequest = Message<"user.v1.LogoutRequest"> & {
};

/**
 * Describes the message user.v1.LogoutRequest.
 * Use `create(LogoutRequestSchema)` to create a new message.
 */
export const LogoutRequestSchema: GenMessage<LogoutRequest> = /*@__PURE__*/
  messageDesc(file_user_v1_auth, 4);

/**
 * @generated from message user.v1.LogoutResponse
 */
export type LogoutResponse = Message<"user.v1.LogoutResponse"> & {
};

/**
 * Describes the message user.v1.LogoutResponse.
 * Use `create(LogoutResponseSchema)` to create a new message.
 */
export const LogoutResponseSchema: GenMessage<LogoutResponse> = /*@__PURE__*/
  messageDesc(file_user_v1_auth, 5);

/**
 * @generated from message user.v1.GetPasskeyIDsRequest
 */
export type GetPasskeyIDsRequest = Message<"user.v1.GetPasskeyIDsRequest"> & {
  /**
   * @generated from field: string username = 1;
   */
  username: string;
};

/**
 * Describes the message user.v1.GetPasskeyIDsRequest.
 * Use `create(GetPasskeyIDsRequestSchema)` to create a new message.
 */
export const GetPasskeyIDsRequestSchema: GenMessage<GetPasskeyIDsRequest> = /*@__PURE__*/
  messageDesc(file_user_v1_auth, 6);

/**
 * @generated from message user.v1.GetPasskeyIDsResponse
 */
export type GetPasskeyIDsResponse = Message<"user.v1.GetPasskeyIDsResponse"> & {
  /**
   * @generated from field: repeated string passkey_ids = 1;
   */
  passkeyIds: string[];
};

/**
 * Describes the message user.v1.GetPasskeyIDsResponse.
 * Use `create(GetPasskeyIDsResponseSchema)` to create a new message.
 */
export const GetPasskeyIDsResponseSchema: GenMessage<GetPasskeyIDsResponse> = /*@__PURE__*/
  messageDesc(file_user_v1_auth, 7);

/**
 * @generated from message user.v1.BeginPasskeyLoginRequest
 */
export type BeginPasskeyLoginRequest = Message<"user.v1.BeginPasskeyLoginRequest"> & {
};

/**
 * Describes the message user.v1.BeginPasskeyLoginRequest.
 * Use `create(BeginPasskeyLoginRequestSchema)` to create a new message.
 */
export const BeginPasskeyLoginRequestSchema: GenMessage<BeginPasskeyLoginRequest> = /*@__PURE__*/
  messageDesc(file_user_v1_auth, 8);

/**
 * @generated from message user.v1.BeginPasskeyLoginResponse
 */
export type BeginPasskeyLoginResponse = Message<"user.v1.BeginPasskeyLoginResponse"> & {
};

/**
 * Describes the message user.v1.BeginPasskeyLoginResponse.
 * Use `create(BeginPasskeyLoginResponseSchema)` to create a new message.
 */
export const BeginPasskeyLoginResponseSchema: GenMessage<BeginPasskeyLoginResponse> = /*@__PURE__*/
  messageDesc(file_user_v1_auth, 9);

/**
 * @generated from message user.v1.FinishPasskeyLoginRequest
 */
export type FinishPasskeyLoginRequest = Message<"user.v1.FinishPasskeyLoginRequest"> & {
};

/**
 * Describes the message user.v1.FinishPasskeyLoginRequest.
 * Use `create(FinishPasskeyLoginRequestSchema)` to create a new message.
 */
export const FinishPasskeyLoginRequestSchema: GenMessage<FinishPasskeyLoginRequest> = /*@__PURE__*/
  messageDesc(file_user_v1_auth, 10);

/**
 * @generated from message user.v1.FinishPasskeyLoginResponse
 */
export type FinishPasskeyLoginResponse = Message<"user.v1.FinishPasskeyLoginResponse"> & {
};

/**
 * Describes the message user.v1.FinishPasskeyLoginResponse.
 * Use `create(FinishPasskeyLoginResponseSchema)` to create a new message.
 */
export const FinishPasskeyLoginResponseSchema: GenMessage<FinishPasskeyLoginResponse> = /*@__PURE__*/
  messageDesc(file_user_v1_auth, 11);

/**
 * @generated from service user.v1.AuthService
 */
export const AuthService: GenService<{
  /**
   * @generated from rpc user.v1.AuthService.Login
   */
  login: {
    methodKind: "unary";
    input: typeof LoginRequestSchema;
    output: typeof LoginResponseSchema;
  },
  /**
   * @generated from rpc user.v1.AuthService.SignUp
   */
  signUp: {
    methodKind: "unary";
    input: typeof SignUpRequestSchema;
    output: typeof SignUpResponseSchema;
  },
  /**
   * @generated from rpc user.v1.AuthService.Logout
   */
  logout: {
    methodKind: "unary";
    input: typeof LogoutRequestSchema;
    output: typeof LogoutResponseSchema;
  },
  /**
   * @generated from rpc user.v1.AuthService.GetPasskeyIDs
   */
  getPasskeyIDs: {
    methodKind: "unary";
    input: typeof GetPasskeyIDsRequestSchema;
    output: typeof GetPasskeyIDsResponseSchema;
  },
  /**
   * @generated from rpc user.v1.AuthService.BeginPasskeyLogin
   */
  beginPasskeyLogin: {
    methodKind: "unary";
    input: typeof BeginPasskeyLoginRequestSchema;
    output: typeof BeginPasskeyLoginResponseSchema;
  },
  /**
   * @generated from rpc user.v1.AuthService.FinishPasskeyLogin
   */
  finishPasskeyLogin: {
    methodKind: "unary";
    input: typeof FinishPasskeyLoginRequestSchema;
    output: typeof FinishPasskeyLoginResponseSchema;
  },
}> = /*@__PURE__*/
  serviceDesc(file_user_v1_auth, 0);

