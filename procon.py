import os
import sys
import time
import json
import signal
import traceback

if sys.platform == "linux":
    from nxbt import Nxbt, PRO_CONTROLLER
else:
    __input__ = """
    {
    "L_STICK": {
        "PRESSED": false,
        "X_VALUE": 0,
        "Y_VALUE": 0,
        "LS_UP": false,
        "LS_LEFT": false,
        "LS_RIGHT": false,
        "LS_DOWN": false
    },
    "R_STICK": {
        "PRESSED": false,
        "X_VALUE": 0,
        "Y_VALUE": 0,
        "RS_UP": false,
        "RS_LEFT": false,
        "RS_RIGHT": false,
        "RS_DOWN": false
    },
    "DPAD_UP": false,
    "DPAD_LEFT": false,
    "DPAD_RIGHT": false,
    "DPAD_DOWN": false,
    "L": false,
    "ZL": false,
    "R": false,
    "ZR": false,
    "JCL_SR": false,
    "JCL_SL": false,
    "JCR_SR": false,
    "JCR_SL": false,
    "PLUS": false,
    "MINUS": false,
    "HOME": false,
    "CAPTURE": false,
    "Y": false,
    "X": false,
    "B": false,
    "A": false
    }
    """

    PRO_CONTROLLER = 3

    class nx:
        state = []

        def __init__(self, **keys):
            self.params = keys
            signal.signal(signal.SIGALRM, self.__alarm)

        def get_available_adapters(self):
            return ["/org/bluez/hci0"]

        def get_switch_addresses(self):
            return ["98:B6:E9:96:E7:FF"]

        def __alarm(self, signum, frame):
            if signum == signal.SIGALRM:
                signal.alarm(0)
                __class__.state[0]["state"] = "crashed"
                __class__.state[0]["errors"] = "error"
            else:
                sys.exit(1)

        def create_controller(
            self,
            controller_type,
            adapter,
            colour_body,
            colour_buttons,
            reconnect_address,
        ):
            if len(self.state) == 0:
                signal.alarm(3)
            __class__.state = [{"state": "created", "errors": None}]
            return 0

        def destroy_controller(self, index):
            __class__.state[index]["state"] = "destroyed"

        def wait_for_connection(self, index):
            time.sleep(1)
            __class__.state[index]["state"] = "connected"

        def create_input_packet(self):
            return json.loads(__input__)

        def set_controller_input(self, index, pkt):
            pass

    def Nxbt(**keys):
        return nx()


sys.stdin = os.fdopen(sys.stdin.fileno(), "r", buffering=1)
sys.stdout = os.fdopen(sys.stdout.fileno(), "w", buffering=1)


class Controller(object):
    def __init__(self) -> None:
        self.nx = None
        self.index = -1

    def close(self) -> None:
        self.disconnect()

    def state(self) -> str:
        if self.index < 0:
            return {"state": "disconnected", "errors": None}
        s = self.nx.state[self.index].copy()
        errors = s["errors"]
        print(type(errors), file=sys.stderr)
        if not errors:
            errors = None
        return {"state": s["state"], "errors": errors}

    def check(self, s=None):
        if self.index < 0:
            return
        state = self.nx.state[self.index].copy()
        if s is not None:
            assert state["state"] == s
        elif state["state"] == "crashed":
            raise Exception(state["errors"].split("\n")[-1])

    def connect(self) -> None:
        self.nx = Nxbt()
        # self.nx.set_debug(True)
        adapters = self.nx.get_available_adapters()
        if len(adapters) < 1:
            raise Exception("not found adapter")
        reconnect_addresses = self.nx.get_switch_addresses()
        self.index = self.nx.create_controller(
            PRO_CONTROLLER,
            adapters[0],
            colour_body=[255, 0, 255],
            colour_buttons=[0, 255, 255],
            reconnect_address=reconnect_addresses,
        )
        self.nx.wait_for_connection(self.index)

    def disconnect(self) -> None:
        if self.index < 0:
            return
        self.nx.remove_controller(self.index)
        self.nx = None
        self.index = -1

    def input(self, pkt) -> None:
        if self.index < 0:
            return
        self.nx.set_controller_input(self.index, pkt)


def main():
    nx = Controller()
    terminate = False
    while terminate is False:
        id = 0
        try:
            line = sys.stdin.readline()
            # print(line, end="", file=sys.stderr)
            req = json.loads(line)
            method = req.get("method", "")
            params = req.get("params", [])
            id = req.get("id", 0)
            result = {}
            if method == "connect":
                # print("connect", file=sys.stderr)
                nx.connect()
            elif method == "disconnect":
                # print("disconnect", file=sys.stderr)
                nx.disconnect()
            elif method == "input":
                # print("input:", params[0], file=sys.stderr)
                nx.input(params[0])
            elif method == "state":
                result = nx.state()
            elif method == "close":
                # print("close", file=sys.stderr)
                terminate = True
            else:
                raise Exception("error: invalid method")
            print(
                json.dumps(
                    {
                        "id": id,
                        "result": result,
                    },
                )
            )
        except Exception as e:
            # traceback.print_exc(file=sys.stderr)
            print()
            print(
                json.dumps(
                    {
                        "id": id,
                        "error": nx.state(),
                    }
                )
            )
        sys.stdout.flush()


if __name__ == "__main__":
    main()
