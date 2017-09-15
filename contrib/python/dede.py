import urllib2
import time
from selenium import webdriver
from selenium.webdriver.common.keys import Keys
from selenium.webdriver.common.desired_capabilities import DesiredCapabilities


class DedeFakeMouse:

    def __init__(self, endpoint, driver):
        self.endpoint = endpoint
        self.driver = driver

    def install(self):
        script = urllib2.urlopen(
            "%s/fake-mouse/install" % self.endpoint).read()
        self.driver.execute_script(script)

    def _fake_mouse_click_on(self, el):
        self.driver.execute_async_script(
            "DedeFakeMouse.clickOn(arguments[0], arguments[1])", el)

    def _fake_mouse_move_on(self, el):
        self.driver.execute_async_script(
            "DedeFakeMouse.moveOn(arguments[0], arguments[1])", el)

    def click_on(self, el):
        self._fake_mouse_click_on(el)
        el.click()

    def move_on(self, el):
        self._fake_mouse_move_on(el)


class DedeTerminal:

    def __init__(self, endpoint, driver):
        self.endpoint = endpoint
        self.driver = driver

    def open_terminal_tab(self, id, width=1400, cols=2000, rows=40, delay=70):
        self.driver.execute_script(
            "window.open('%s/terminal/%s?width=%d&cols=%d&rows=%d&delay=%d')" %
            (self.endpoint, id, width, cols, rows, delay))
        self.driver.switch_to_window(self.driver.window_handles[-1])

    def start_record(self):
        self.driver.execute_script("DedeTerminal.startRecord()")

    def stop_record(self):
        self.driver.execute_script("DedeTerminal.stopRecord()")

    def type(self, str):
        self.driver.execute_async_script(
            "DedeTerminal.type(arguments[0], arguments[1])", str)

    def type_cmd(self, str):
        self.driver.execute_async_script(
            "DedeTerminal.typeCmd(arguments[0], arguments[1])", str)

    def type_cmd_wait(self, str, regex):
        self.driver.execute_async_script(
            "DedeTerminal.typeCmdWait(arguments[0], arguments[1], arguments[2])", str, regex)

driver = webdriver.Remote(
  command_executor='http://127.0.0.1:4444/wd/hub',
  desired_capabilities={"browserName": "chrome"})
driver.maximize_window()
driver.get("https://github.com/skydive-project/dede")
driver.set_script_timeout(5)

time.sleep(2)

dede = DedeFakeMouse("http://localhost:12345", driver)
dede.install()

time.sleep(2)

# start the demo
clone = driver.find_element_by_xpath(
    "//button[@aria-label='Clone or download this repository']")
dede.click_on(clone)
copy = driver.find_element_by_xpath(
    "//button[@aria-label='Copy to clipboard']")
dede.click_on(copy)
input = driver.find_element_by_xpath(
    "//input[contains(@aria-label, 'Clone this repository at')]")
url = input.get_property("value")

time.sleep(1)

terminal = DedeTerminal("http://192.168.1.21:12345", driver)
terminal.open_terminal_tab("test1")

time.sleep(1)

terminal.type_cmd_wait("cd /tmp", "safchain")
terminal.type_cmd_wait("git clone %s" % url, "safchain")
terminal.type_cmd_wait("cd dede", "safchain")

time.sleep(10)

driver.close()
