import { describe, it, expect } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";
import Dropdown from "./Dropdown";

const options = [
  { value: "", label: "All" },
  { value: "1", label: "Option 1" },
  { value: "2", label: "Option 2" },
];

describe("Dropdown", () => {
  it("renders with placeholder when no value matches", () => {
    const opts = [
      { value: "1", label: "Option 1" },
      { value: "2", label: "Option 2" },
    ];
    render(<Dropdown value="" options={opts} onChange={() => {}} placeholder="Pick one" />);
    expect(screen.getByText("Pick one")).toBeInTheDocument();
  });

  it("renders selected option label", () => {
    render(<Dropdown value="1" options={options} onChange={() => {}} />);
    expect(screen.getByText("Option 1")).toBeInTheDocument();
  });

  it("opens and shows options on click", () => {
    render(<Dropdown value="" options={options} onChange={() => {}} />);
    fireEvent.click(screen.getByRole("button"));
    expect(screen.getByText("Option 1")).toBeInTheDocument();
    expect(screen.getByText("Option 2")).toBeInTheDocument();
  });

  it("calls onChange when option is clicked", () => {
    let selected = "";
    render(<Dropdown value="" options={options} onChange={(v) => { selected = v; }} />);
    fireEvent.click(screen.getByRole("button"));
    fireEvent.click(screen.getByText("Option 2"));
    expect(selected).toBe("2");
  });

  it("closes after selecting an option", () => {
    render(<Dropdown value="" options={options} onChange={() => {}} />);
    fireEvent.click(screen.getByRole("button"));
    // Options list is open — Option 2 appears in the list
    expect(screen.getByText("Option 2")).toBeInTheDocument();
    fireEvent.click(screen.getByText("Option 2"));
    // After selecting, the dropdown list should close
    // The trigger button now shows "All" (value="" matched), and the list is gone
    expect(screen.queryAllByRole("button")).toHaveLength(1);
  });

  it("does not open when disabled", () => {
    render(<Dropdown value="" options={options} onChange={() => {}} disabled />);
    fireEvent.click(screen.getByRole("button"));
    // Options should not appear
    expect(screen.queryByText("Option 1")).not.toBeInTheDocument();
  });
});
