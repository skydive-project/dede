import urllib2
import time
from selenium import webdriver
from selenium.webdriver.common.keys import Keys
from selenium.webdriver.common.desired_capabilities import DesiredCapabilities


class DedeChapterManager:

    def __init__(self, dede, chapterID):
        self.dede = dede
        self.chapterID = chapterID

    def __enter__(self):
        self.prevChapterID = self.dede.chapterID
        self.dede.chapterID = self.chapterID

    def __exit__(self, type, value, traceback):
        self.dede.chapterID = self.prevChapterID


class DedeSectionManager:

    def __init__(self, dede, sectionID):
        self.dede = dede
        self.sectionID = sectionID

    def __enter__(self):
        self.prevSectionID = self.dede.sectionID
        self.dede.sectionID = self.sectionID

    def __exit__(self, type, value, traceback):
        self.dede.sectionID = self.prevSectionID


class Dede:

    def __init__(self, endpoint, driver, sessionID):
        self.endpoint = endpoint
        self.driver = driver
        self.sessionID = sessionID
        self.chapterID = ''
        self.sectionID = ''

    def fake_mouse(self):
        return DedeFakeMouse(self)

    def terminal_manager(self):
        return DedeTerminalManager(self)

    def video_recorder(self):
        return DedeVideoRecorder(self)

    def chapter(self, chapterID):
        return DedeChapterManager(self, chapterID)

    def section(self, sectionID):
        return DedeSectionManager(self, sectionID)


class DedeFakeMouse:

    def __init__(self, dede):
        self.dede = dede

    def install(self):
        # TODO catch error
        print("%s/fake-mouse/install" % self.dede.endpoint)
        script = urllib2.urlopen(
            "%s/fake-mouse/install" % self.dede.endpoint).read()
        self.dede.driver.execute_script(script)

    def _fake_mouse_click_on(self, el):
        self.dede.driver.execute_async_script(
            "DedeFakeMouse.clickOn(arguments[0], arguments[1])", el)

    def _fake_mouse_move_on(self, el):
        self.dede.driver.execute_async_script(
            "DedeFakeMouse.moveOn(arguments[0], arguments[1])", el)

    def click_on(self, el):
        self._fake_mouse_click_on(el)
        el.click()

    def move_on(self, el):
        self._fake_mouse_move_on(el)


class DedeTerminalManagerTab:

    def __init__(self, dede, window_handle):
        self.dede = dede
        self.window_handle = window_handle

    def focus(self):
        self.dede.driver.switch_to_window(self.window_handle)

    def start_record(self):
        self.dede.driver.execute_script(
            "DedeTerminal.startRecord(%d, %d, %d)" %
            (self.dede.sessionID, self.dede.chapterID, self.dede.sectionID))

    def stop_record(self):
        self.dede.driver.execute_script("DedeTerminal.stopRecord()")

    def type(self, str):
        self.dede.driver.execute_async_script(
            "DedeTerminal.type(arguments[0], arguments[1])", str)

    def type_cmd(self, str):
        self.dede.driver.execute_async_script(
            "DedeTerminal.typeCmd(arguments[0], arguments[1])", str)

    def type_cmd_wait(self, str, regex):
        self.dede.driver.execute_async_script(
            "DedeTerminal.typeCmdWait("
            "arguments[0], arguments[1], arguments[2])", str, regex)


class DedeTerminalManager:

    def __init__(self, dede):
        self.dede = dede
        self.termIndex = 1

    def open_terminal_tab(
            self, title, width=1400, cols=2000, rows=40, delay=70):
        self.dede.driver.execute_script(
            "window.open('%s/terminal/%s?"
            "title=%s&width=%d&cols=%d&rows=%d&delay=%d')" %
            (self.dede.endpoint, self.termIndex,
             title, width, cols, rows, delay))
        self.termIndex += 1

        window_handle = self.dede.driver.window_handles[-1]
        tab = DedeTerminalManagerTab(self.dede, window_handle)
        self.dede.driver.switch_to_window(window_handle)

        return tab


class DedeVideoRecord:

    def __init__(self, dede):
        self.dede = dede

    def stop(self):
        # TODO catch error
        urllib2.urlopen(
            "%s/video/stop-record?sessionID=%s&chapterID=%s&sectionID=%s" %
            (self.dede.endpoint, self.dede.sessionID,
             self.dede.chapterID, self.dede.sectionID))


class DedeVideoRecorder:

    def __init__(self, dede):
        self.dede = dede

    def start_record(self):
        # TODO catch error
        urllib2.urlopen(
            "%s/video/start-record?sessionID=%s&chapterID=%s&sectionID=%s" %
            (self.dede.endpoint, self.dede.sessionID,
             self.dede.chapterID, self.dede.sectionID))
        return DedeVideoRecord(self.dede)


if __name__ == '__main__':
    driver = webdriver.Remote(
      command_executor='http://127.0.0.1:4444/wd/hub',
      desired_capabilities={"browserName": "chrome"})
    driver.maximize_window()
    driver.get("https://github.com/skydive-project/dede")
    driver.set_script_timeout(20)

    time.sleep(2)

    dede = Dede("http://192.168.1.21:55555", driver, 1)
    fake_mouse = dede.fake_mouse()
    fake_mouse.install()

    with dede.chapter(1):

        record = dede.video_recorder().start_record()

        time.sleep(2)

        # start the demo
        clone = driver.find_element_by_xpath(
            "//details[contains(@class, 'get-repo-select-menu')]")
        fake_mouse.click_on(clone)

        copy = driver.find_element_by_xpath(
            "//button[@aria-label='Copy to clipboard']")
        fake_mouse.click_on(copy)

        input = driver.find_element_by_xpath(
            "//input[contains(@aria-label, 'Clone this repository at')]")
        url = input.get_property("value")

        time.sleep(1)

        tab1 = dede.terminal_manager().open_terminal_tab('clone')
        tab2 = dede.terminal_manager().open_terminal_tab('list')

        time.sleep(1)

        with dede.section(1):
            tab1.focus()
            tab1.start_record()
            tab1.type_cmd_wait("cd /tmp", "safchain")
            tab1.type_cmd_wait("git clone %s" % url, "safchain")
            tab1.type_cmd_wait("cd dede", "safchain")
            tab1.stop_record()

        with dede.section(2):
            tab2.focus()
            tab2.start_record()
            tab2.type_cmd_wait("ls -al", "safchain")
            tab2.stop_record()

        record.stop()

    time.sleep(10)

    driver.close()
