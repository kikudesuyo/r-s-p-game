from util import display_result, get_winner, display_input

value = None
while value == None:
  hand_list = display_input()
  player_hand = hand_list[0]
  opponent_hand = hand_list[1]
  value = get_winner(player_hand, opponent_hand)
  display_result(value)
