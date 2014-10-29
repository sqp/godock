
// #include <signal.h>

#include "cairo-dock-dock-manager.h"       // myDocksParam
#include "cairo-dock-desklet-manager.h"          // myDeskletObjectMgr
#include "cairo-dock-keybinder.h"                // myShortkeyObjectMgr
#include "cairo-dock-module-instance-manager.h"  // myModuleObjectMgr
#include "cairo-dock-module-manager.h"           // myModuleObjectMgr

#include "gldi-config.h"                         // GLDI_VERSION

#include "cairo-dock-gui-manager.h"              // CairoDockGuiBackend

// #include <unistd.h> // sleep, execl
// #define __USE_POSIX
// #include <time.h>

// #include <glib/gstdio.h>
// #include "cairo-dock-icon-facility.h"  // cairo_dock_get_first_icon
// #include "cairo-dock-themes-manager.h"
// #include "cairo-dock-dialog-factory.h"
// #include "cairo-dock-keyfile-utilities.h"
// #include "cairo-dock-file-manager.h"
// #include "cairo-dock-packages.h"

// #include "cairo-dock-config.h"
// #include "cairo-dock-log.h"
// #include "cairo-dock-utils.h"  // cairo_dock_launch_command
// #include "cairo-dock-core.h"


// local files
#include "cairo-dock-gui-backend.h"
#include "cairo-dock-user-menu.h"
#include "cairo-dock-user-interaction.h"


extern int g_iMajorVersion, g_iMinorVersion, g_iMicroVersion;

extern gchar *g_cCairoDockDataDir;
extern gchar *g_cCurrentThemePath;

extern gchar *g_cConfFile;

extern CairoDock *g_pMainDock;



// extern gboolean g_bUseOpenGL;
// extern gboolean g_bEasterEggs;

// extern GldiModuleInstance *g_pCurrentModule;


// TODO: To remove once the menu is redone:

gboolean g_bLocked;
#define CAIRO_DOCK_VERSION GLDI_VERSION  // using GLDI_VERSION instead (remove once menu is redone)


// Those that may have to be reimplemented:

// static gint s_iNbCrashes = 0;
// static gboolean s_bSucessfulLaunch = FALSE;
// static GString *s_pLaunchCommand = NULL;
// static gint s_iLastYear = 0;
// static gboolean s_bCDSessionLaunched = FALSE; // session CD already launched?

// static gchar *s_cLastVersion = NULL;
// static gchar *s_cDefaulBackend = NULL;


// GUI CALLBACK BACKEND

extern GtkWidget* ShowMainGui           ();
extern GtkWidget* ShowModuleGui         (gchar *cModuleName);
extern GtkWidget *ShowGui               (Icon *pIcon, GldiContainer *pContainer, GldiModuleInstance *pModuleInstance, int iShowPage);

extern GtkWidget * ShowAddons();
extern void ReloadItems();
extern void Reload();
extern void Close();
extern void UpdateModulesList();
extern void UpdateModuleState(gchar*, gboolean);
extern void UpdateModuleInstanceContainer(GldiModuleInstance*, gboolean);
extern void UpdateShortkeys();

// GUI CORE BACKEND
extern void ShowModuleInstanceGui (GldiModuleInstance *pModuleInstance, int iShowPage);
extern void SetStatusMessage (gchar *cMessage);
extern void ReloadCurrentWidget (GldiModuleInstance *pInstance, int iShowPage);



// WRAPPERS AROUND CALLS IMPOSSIBLE FROM C TO GO (const gchar* -> gchar*)

static GtkWidget* _showModuleGui (const gchar *cModuleName) { return ShowModuleGui(g_strdup(cModuleName)); }
static void       _updateModuleState (const gchar *cModuleName, gboolean bActive) { UpdateModuleState(g_strdup(cModuleName), bActive); }
static void       _setStatusMessage (const gchar *message) { SetStatusMessage(g_strdup(message)); }


static void register_gui (void)
{
	CairoDockMainGuiBackend *pBackend = g_new0 (CairoDockMainGuiBackend, 1);
	cairo_dock_register_config_gui_backend (pBackend);

 	pBackend->show_main_gui 					= ShowMainGui;
 	pBackend->show_module_gui 					= _showModuleGui;
 	pBackend->show_gui 							= ShowGui;

	pBackend->close_gui 						= Close;
	pBackend->update_module_state 				= _updateModuleState;
	pBackend->update_module_instance_container 	= UpdateModuleInstanceContainer;
// 	pBackend->update_desklet_params 			= update_desklet_params;
// 	pBackend->update_desklet_visibility_params 	= update_desklet_visibility_params;
	pBackend->update_modules_list 				= UpdateModulesList;
	pBackend->update_shortkeys 					= UpdateShortkeys;
	pBackend->show_addons 						= ShowAddons;
	pBackend->reload_items 						= ReloadItems;
	pBackend->reload 							= Reload;
// 	pBackend->cDisplayedName 					= _("Advanced Mode");  // name of the other backend.
// 	pBackend->cTooltip 							= _("The advanced mode lets you tweak every single parameter of the dock. It is a powerful tool to customise your current theme.");


 	CairoDockGuiBackend *pConfigBackend = g_new0 (CairoDockGuiBackend, 1);

	pConfigBackend->set_status_message_on_gui 	= _setStatusMessage;
	pConfigBackend->reload_current_widget 		= ReloadCurrentWidget;
 	pConfigBackend->show_module_instance_gui 	= ShowModuleInstanceGui;
// 	pConfigBackend->get_widget_from_name 		= get_widget_from_name;

 	cairo_dock_register_gui_backend (pConfigBackend);
}


// UNCHANGED
static void register_events() {
	//\___________________ register to the useful notifications.
	gldi_object_register_notification (&myContainerObjectMgr,
		NOTIFICATION_DROP_DATA,
		(GldiNotificationFunc) cairo_dock_notification_drop_data,
		GLDI_RUN_AFTER, NULL);
	gldi_object_register_notification (&myContainerObjectMgr,
		NOTIFICATION_CLICK_ICON,
		(GldiNotificationFunc) cairo_dock_notification_click_icon,
		GLDI_RUN_AFTER, NULL);
	gldi_object_register_notification (&myContainerObjectMgr,
		NOTIFICATION_MIDDLE_CLICK_ICON,
		(GldiNotificationFunc) cairo_dock_notification_middle_click_icon,
		GLDI_RUN_AFTER, NULL);
	gldi_object_register_notification (&myContainerObjectMgr,
		NOTIFICATION_SCROLL_ICON,
		(GldiNotificationFunc) cairo_dock_notification_scroll_icon,
		GLDI_RUN_AFTER, NULL);
	gldi_object_register_notification (&myContainerObjectMgr,
		NOTIFICATION_BUILD_CONTAINER_MENU,
		(GldiNotificationFunc) cairo_dock_notification_build_container_menu,
		GLDI_RUN_FIRST, NULL);
	gldi_object_register_notification (&myContainerObjectMgr,
		NOTIFICATION_BUILD_ICON_MENU,
		(GldiNotificationFunc) cairo_dock_notification_build_icon_menu,
		GLDI_RUN_AFTER, NULL);
	
	gldi_object_register_notification (&myDeskletObjectMgr,
		NOTIFICATION_CONFIGURE_DESKLET,
		(GldiNotificationFunc) cairo_dock_notification_configure_desklet,
		GLDI_RUN_AFTER, NULL);
	gldi_object_register_notification (&myDockObjectMgr,
		NOTIFICATION_ICON_MOVED,
		(GldiNotificationFunc) cairo_dock_notification_icon_moved,
		GLDI_RUN_AFTER, NULL);
	gldi_object_register_notification (&myDockObjectMgr,
		NOTIFICATION_DESTROY,
		(GldiNotificationFunc) cairo_dock_notification_dock_destroyed,
		GLDI_RUN_AFTER, NULL);
	gldi_object_register_notification (&myModuleObjectMgr,
		NOTIFICATION_MODULE_ACTIVATED,
		(GldiNotificationFunc) cairo_dock_notification_module_activated,
		GLDI_RUN_AFTER, NULL);
	gldi_object_register_notification (&myModuleObjectMgr,
		NOTIFICATION_MODULE_REGISTERED,
		(GldiNotificationFunc) cairo_dock_notification_module_registered,
		GLDI_RUN_AFTER, NULL);
	gldi_object_register_notification (&myModuleInstanceObjectMgr,
		NOTIFICATION_MODULE_INSTANCE_DETACHED,
		(GldiNotificationFunc) cairo_dock_notification_module_detached,
		GLDI_RUN_AFTER, NULL);
	gldi_object_register_notification (&myDockObjectMgr,
		NOTIFICATION_INSERT_ICON,
		(GldiNotificationFunc) cairo_dock_notification_icon_inserted,
		GLDI_RUN_AFTER, NULL);
	gldi_object_register_notification (&myDockObjectMgr,
		NOTIFICATION_REMOVE_ICON,
		(GldiNotificationFunc) cairo_dock_notification_icon_removed,
		GLDI_RUN_AFTER, NULL);
	gldi_object_register_notification (&myDeskletObjectMgr,
		NOTIFICATION_DESTROY,
		(GldiNotificationFunc) cairo_dock_notification_desklet_added_removed,
		GLDI_RUN_AFTER, NULL);
	gldi_object_register_notification (&myDeskletObjectMgr,
		NOTIFICATION_NEW,
		(GldiNotificationFunc) cairo_dock_notification_desklet_added_removed,
		GLDI_RUN_AFTER, NULL);
	gldi_object_register_notification (&myShortkeyObjectMgr,
		NOTIFICATION_NEW,
		(GldiNotificationFunc) cairo_dock_notification_shortkey_added_removed_changed,
		GLDI_RUN_AFTER, NULL);
	gldi_object_register_notification (&myShortkeyObjectMgr,
		NOTIFICATION_DESTROY,
		(GldiNotificationFunc) cairo_dock_notification_shortkey_added_removed_changed,
		GLDI_RUN_AFTER, NULL);
	gldi_object_register_notification (&myShortkeyObjectMgr,
		NOTIFICATION_SHORTKEY_CHANGED,
		(GldiNotificationFunc) cairo_dock_notification_shortkey_added_removed_changed,
		GLDI_RUN_AFTER, NULL);
}




// may still need to check s_iLastYear and s_bCDSessionLaunched 
/*
// UNCHANGED
static void _cairo_dock_get_global_config (const gchar *cCairoDockDataDir)
{
	gchar *cConfFilePath = g_strdup_printf ("%s/.cairo-dock", cCairoDockDataDir);
	GKeyFile *pKeyFile = g_key_file_new ();
	if (g_file_test (cConfFilePath, G_FILE_TEST_EXISTS))
	{
		g_key_file_load_from_file (pKeyFile, cConfFilePath, 0, NULL);
		s_cLastVersion = g_key_file_get_string (pKeyFile, "Launch", "last version", NULL);
		s_cDefaulBackend = g_key_file_get_string (pKeyFile, "Launch", "default backend", NULL);
		if (s_cDefaulBackend && *s_cDefaulBackend == '\0')
		{
			g_free (s_cDefaulBackend);
			s_cDefaulBackend = NULL;
		}
		// s_iGuiMode = g_key_file_get_integer (pKeyFile, "Gui", "mode", NULL);  // 0 by default
		s_iLastYear = g_key_file_get_integer (pKeyFile, "Launch", "last year", NULL);  // 0 by default
		// s_bPingServer = g_key_file_get_boolean (pKeyFile, "Launch", "ping server", NULL);  // FALSE by default
		s_bCDSessionLaunched = g_key_file_get_boolean (pKeyFile, "Launch", "cd session", NULL);  // FALSE by default
	}
	else  // first launch or old version, the file doesn't exist yet.
	{
		gchar *cLastVersionFilePath = g_strdup_printf ("%s/.cairo-dock-last-version", cCairoDockDataDir);
		if (g_file_test (cLastVersionFilePath, G_FILE_TEST_EXISTS))
		{
			gsize length = 0;
			g_file_get_contents (cLastVersionFilePath,
				&s_cLastVersion,
				&length,
				NULL);
		}
		g_remove (cLastVersionFilePath);
		g_free (cLastVersionFilePath);
		g_key_file_set_string (pKeyFile, "Launch", "last version", s_cLastVersion?s_cLastVersion:"");
		
		g_key_file_set_string (pKeyFile, "Launch", "default backend", "");
		
		// g_key_file_set_integer (pKeyFile, "Gui", "mode", s_iGuiMode);
		
		g_key_file_set_integer (pKeyFile, "Launch", "last year", s_iLastYear);
		
		cairo_dock_write_keys_to_file (pKeyFile, cConfFilePath);
	}
	g_key_file_free (pKeyFile);
	g_free (cConfFilePath);
}
*/
