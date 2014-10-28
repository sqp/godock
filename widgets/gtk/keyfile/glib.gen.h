#pragma once
#include <stdlib.h>
#include <stdint.h>

typedef struct _GError GError;

typedef struct _GKeyFile GKeyFile;
typedef uint32_t GKeyFileError;
typedef uint32_t GKeyFileFlags;

extern GError* g_error_new_literal(uint32_t, int32_t, char*);
extern GError* g_error_copy(GError*);
extern void g_error_free(GError*);
extern int g_error_matches(GError*, uint32_t, int32_t);

extern GKeyFile* g_key_file_new();
extern int g_key_file_get_boolean(GKeyFile*, char*, char*, GError**);
extern int* g_key_file_get_boolean_list(GKeyFile*, char*, char*, uint64_t*, GError**);
extern char* g_key_file_get_comment(GKeyFile*, char*, char*, GError**);
extern double g_key_file_get_double(GKeyFile*, char*, char*, GError**);
extern double* g_key_file_get_double_list(GKeyFile*, char*, char*, uint64_t*, GError**);
extern char** g_key_file_get_groups(GKeyFile*, uint64_t*);
extern int64_t g_key_file_get_int64(GKeyFile*, char*, char*, GError**);
extern int32_t g_key_file_get_integer(GKeyFile*, char*, char*, GError**);
extern int32_t* g_key_file_get_integer_list(GKeyFile*, char*, char*, uint64_t*, GError**);
extern char** g_key_file_get_keys(GKeyFile*, char*, uint64_t*, GError**);
extern char* g_key_file_get_locale_string(GKeyFile*, char*, char*, char*, GError**);
extern char** g_key_file_get_locale_string_list(GKeyFile*, char*, char*, char*, uint64_t*, GError**);
extern char* g_key_file_get_start_group(GKeyFile*);
extern char* g_key_file_get_string(GKeyFile*, char*, char*, GError**);
extern char** g_key_file_get_string_list(GKeyFile*, char*, char*, uint64_t*, GError**);
extern uint64_t g_key_file_get_uint64(GKeyFile*, char*, char*, GError**);
extern char* g_key_file_get_value(GKeyFile*, char*, char*, GError**);
extern int g_key_file_has_group(GKeyFile*, char*);
extern int g_key_file_load_from_data(GKeyFile*, char*, uint64_t, GKeyFileFlags, GError**);
extern int g_key_file_load_from_data_dirs(GKeyFile*, char*, char**, GKeyFileFlags, GError**);
extern int g_key_file_load_from_dirs(GKeyFile*, char*, char**, char**, GKeyFileFlags, GError**);
extern int g_key_file_load_from_file(GKeyFile*, char*, GKeyFileFlags, GError**);
extern int g_key_file_remove_comment(GKeyFile*, char*, char*, GError**);
extern int g_key_file_remove_group(GKeyFile*, char*, GError**);
extern int g_key_file_remove_key(GKeyFile*, char*, char*, GError**);
extern void g_key_file_set_boolean(GKeyFile*, char*, char*, int);
extern void g_key_file_set_boolean_list(GKeyFile*, char*, char*, int*, uint64_t);
extern int g_key_file_set_comment(GKeyFile*, char*, char*, char*, GError**);
extern void g_key_file_set_double(GKeyFile*, char*, char*, double);
extern void g_key_file_set_double_list(GKeyFile*, char*, char*, double*, uint64_t);
extern void g_key_file_set_int64(GKeyFile*, char*, char*, int64_t);
extern void g_key_file_set_integer(GKeyFile*, char*, char*, int32_t);
extern void g_key_file_set_integer_list(GKeyFile*, char*, char*, int32_t*, uint64_t);
extern void g_key_file_set_list_separator(GKeyFile*, uint8_t);
extern void g_key_file_set_locale_string(GKeyFile*, char*, char*, char*, char*);
extern void g_key_file_set_locale_string_list(GKeyFile*, char*, char*, char*, char**, uint64_t);
extern void g_key_file_set_string(GKeyFile*, char*, char*, char*);
extern void g_key_file_set_string_list(GKeyFile*, char*, char*, char**, uint64_t);
extern void g_key_file_set_uint64(GKeyFile*, char*, char*, uint64_t);
extern void g_key_file_set_value(GKeyFile*, char*, char*, char*);
extern char* g_key_file_to_data(GKeyFile*, uint64_t*, GError**);
extern void g_key_file_unref(GKeyFile*);
extern uint32_t g_key_file_error_quark();

extern void g_free(void*);
